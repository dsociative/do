package do

import (
	"database/sql/driver"
	"fmt"
	"reflect"

	"log"

	"github.com/fatih/structs"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
)

var (
	ptInt64Conv  = pointerConverter[int64]("PtInt64")
	ptStringConv = pointerConverter[string]("PtString")
	strSliceConv = sliceConverter[string]("StrSlice")
)

func ValuerConv() cmp.Option {
	return cmp.FilterValues(
		func(x, y any) bool {
			comparable := func(v interface{}) bool {
				switch v.(type) {
				case driver.Valuer:
					return true
				default:
					return false
				}
			}
			return comparable(x) && comparable(y)
		},
		cmp.Transformer(
			"valuer",
			func(i any) any {
				switch v := i.(type) {
				case driver.Valuer:
					value, err := v.Value()
					if err != nil {
						log.Panicf("valuer error %v - %v", i, err)
					}
					return value
				default:
					return i
				}
			},
		),
	)
}

func pointerConverter[T any](name string) cmp.Option {
	return cmp.FilterValues(func(x, y interface{}) bool {
		comparable := func(v interface{}) (hasPointer bool, ok bool) {
			switch v.(type) {
			case *T:
				hasPointer = true
				ok = true
				return
			case T:
				ok = true
			}
			return
		}
		hasPtX, okX := comparable(x)
		hasPtY, okY := comparable(y)
		return (hasPtX || hasPtY) && (okX && okY)
	},
		cmp.Transformer(name, func(i interface{}) (noValue T) {
			switch v := i.(type) {
			case *T:
				return UnPt(v)
			case T:
				return v
			default:
				return noValue
			}
		}),
	)
}

func sliceConverter[T any](name string) cmp.Option {
	var slice []T
	t := reflect.TypeOf(slice)
	return cmp.FilterValues(func(x, y interface{}) bool {
		if t == nil {
			log.Fatal("slice converter reflect nil")
		}
		comparable := func(v interface{}) bool {
			if v == nil {
				return false
			}
			return reflect.TypeOf(v).AssignableTo(t)
		}
		return comparable(x) && comparable(y)
	},
		cmp.Transformer(name, func(i interface{}) []T {
			v := reflect.ValueOf(i)
			n := v.Len()
			s := reflect.MakeSlice(t, n, n)
			reflect.Copy(s, v)
			return s.Interface().([]T)
		}),
	)
}

func CmpOnlyStructNames(s any) cmp.Option {
	return CmpOnly(structs.Names(s)...)
}

func CmpOnly(fields ...string) cmp.Option {
	return cmp.FilterPath(mapFieldPath(func(field string) bool {
		return !Contains(fields, field)
	}), cmp.Ignore())
}

func mapFieldPath(fn func(key string) bool, fields ...string) func(p cmp.Path) bool {
	t := reflect.TypeOf(Key(""))
	return func(p cmp.Path) bool {
		switch ps := p.Last().(type) {
		case cmp.MapIndex:
			if ps.Key().Type() != t {
				return false
			}
			return fn(ps.Key().String())
		default:
			return false
		}
	}
}

func CmpIgnore(fields ...string) cmp.Option {
	return cmp.FilterPath(mapFieldPath(func(field string) bool {
		return Contains(fields, field)
	}), cmp.Ignore())
}

func CmpKeyTransform[In any, Out any](field string, cv func(In) Out) cmp.Option {
	return cmp.FilterPath(
		mapFieldPath(func(key string) bool {
			return key == field
		}),
		cmpopts.AcyclicTransformer("keyTrans"+field, func(i any) any {
			if typedValue, ok := i.(In); ok {
				return cv(typedValue)
			}
			return i
		}),
	)
}

func CmpKeyComparer[X any, Y any](field string, cmpFn func(X, Y) bool) cmp.Option {
	return cmp.FilterPath(
		mapFieldPath(func(key string) bool {
			fmt.Println(key, key == field)
			return key == field
		}),
		cmp.Comparer(func(x, y any) bool {
			if tx, ok := x.(X); ok {
				if ty, ok := y.(Y); ok {
					return cmpFn(tx, ty)
				}
			}
			return true
		}),
	)
}

type RenameMap map[string]string

func CmpRename(renameMap RenameMap) cmp.Option {
	return cmp.Transformer("cmpRename", func(m Untyped) Untyped {
		for f, newF := range renameMap {
			field := Key(f)
			newField := Key(newF)
			if f, ok := m[field]; ok {
				_, hasAlready := m[newField]
				if hasAlready {
					log.Fatalf("cmp rename '%s' to already existed field '%s'", field, newField)
				}

				m[newField] = f
				delete(m, field)
			}
		}
		return m
	})
}

type Key string
type Untyped map[Key]any

func SliceCmp[E any, R any](t require.TestingT, expected []E, real []R, opts ...cmp.Option) {
	opts = append(opts, ptInt64Conv, ptStringConv)
	var (
		e, r []Untyped
	)
	e = Map(expected, toMap[E])
	r = Map(real, toMap[R])
	require.Empty(t, cmp.Diff(e, r, opts...))
}

func MappedCmp[E any, R any](t require.TestingT, expectedIDField, realIDField string, expected []E, real []R, opts ...cmp.Option) {
	opts = append(opts, ptInt64Conv, ptStringConv, strSliceConv)
	MappedCmpStrict(t, expectedIDField, realIDField, expected, real, opts...)
}

func MappedCmpStrict[E any, R any](t require.TestingT, expectedIDField, realIDField string, expected []E, real []R, opts ...cmp.Option) {
	e, err := toMapped(expectedIDField, expected)
	require.NoError(t, err)
	r, err := toMapped(realIDField, real)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(e, r, opts...))
}

func toMapped[T any](idName string, items []T) (map[interface{}]Untyped, error) {
	values := Map(items, toMap[T])
	idField := Key(idName)
	m := map[interface{}]Untyped{}
	for _, v := range values {
		id, ok := v[idField]
		if !ok {
			return nil, fmt.Errorf("no '%s' field in %+v ", idField, v)
		}

		id = toString[string](id)
		_, alreadyExist := m[id]
		if alreadyExist {
			return nil, fmt.Errorf("duplicate '%v'", id)
		}

		m[id] = v
		delete(v, idField)
	}
	return m, nil
}

func toString[T any](i interface{}) interface{} {
	switch v := i.(type) {
	case *T:
		return UnPt(v)
	case T:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toMap[V any](i V) Untyped {
	m := Untyped{}
	for k, v := range structs.New(i).Map() {
		m[Key(k)] = v
	}
	return m
}

type TestingT interface {
	require.TestingT
	Helper()
}

func UntypedCmp[E any, R any](t TestingT, expected E, real R, opts ...cmp.Option) {
	t.Helper()
	e := ToUntyped(expected)
	r := ToUntyped(real)
	require.Empty(t, cmp.Diff(e, r, opts...))
}

func ToUntyped[V any](i V) Untyped {
	m := Untyped{}

	v := reflect.ValueOf(i)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		a := f.Interface()
		m[Key(v.Type().Field(i).Name)] = a
	}
	return m
}

package do

func Map[In any, Out any](items []In, fn func(In) Out) []Out {
	out := make([]Out, 0, len(items))
	for _, i := range items {
		out = append(out, fn(i))
	}
	return out
}

func Group[In any, Out comparable](items []In, fn func(In) Out) map[Out][]In {
	out := map[Out][]In{}
	for _, i := range items {
		o := fn(i)
		if _, ok := out[o]; ok {
			out[o] = []In{}
		}
		out[o] = append(out[o], i)
	}
	return out
}

func Find[T any](items []T, fn func(T) bool) (r T, _ bool) {
	for _, i := range items {
		if fn(i) {
			return i, true
		}
	}
	return r, false
}

func Filter[T any](items []T, fn func(T) bool) (r []T) {
	for _, i := range items {
		if fn(i) {
			r = append(r, i)
		}
	}
	return
}

func Split[T any](items []T, fn func(T) bool) (trueValues, falseValues []T) {
	for _, i := range items {
		if fn(i) {
			trueValues = append(trueValues, i)
		} else {
			falseValues = append(falseValues, i)
		}
	}
	return
}

func Count[T any](items []T, fn func(T) bool) int {
	return len(Filter(items, fn))
}

func PopFind[T any](items []T, fn func(T) bool) (r T, _ bool, out []T) {
	for i, item := range items {
		if fn(item) {
			return item, true, append(items[:i], items[i+1:]...)
		}
	}
	return r, false, items
}

func Flatten[T any](items ...[]T) (out []T) {
	for _, i := range items {
		out = append(out, i...)
	}
	return
}

func FlattenMap[In any, Out any](items []In, fn func(In) []Out) []Out {
	return Flatten(Map(items, fn)...)
}

func Contains[T comparable](items []T, v ...T) bool {
	for _, i := range items {
		for _, b := range v {
			if i == b {
				return true
			}
		}
	}
	return false
}

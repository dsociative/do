package do

import (
	"log"
	"time"

	"github.com/go-openapi/strfmt"
)

func Pt[T any](v T) *T {
	return &v
}

func UnPt[T any](v *T) (dv T) {
	if v == nil {
		return dv
	}
	return *v
}

func PtMult[T int64 | int | float64](pt *T, mult T) *T {
	return Pt(UnPt(pt) * mult)
}

func PtEq[T comparable](a, b *T) bool {
	return UnPt(a) == UnPt(b)
}

func PtStrTime(s *string) (time.Time, error) {
	return time.Parse(time.RFC3339, UnPt(s))
}

func PtStrfTime(s *strfmt.DateTime) time.Time {
	return time.Time(UnPt(s))
}

func PtStrAfter(now time.Time, s *string) bool {
	if s == nil {
		return false
	}
	t, err := PtStrTime(s)
	if err != nil {
		log.Fatalf("can't parse time: %+v %s", s, err)
	}
	return t.After(now)
}

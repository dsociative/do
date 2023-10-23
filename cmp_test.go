package do

import (
	"testing"
)

type V1 struct {
	FieldOne string
}

type V2 struct {
	FieldOne string
	FieldTwo string
}

type V3 struct {
	FieldElse string
	FieldTwo  string
}

func TestMappedCmp(t *testing.T) {
	SliceCmp(
		t,
		[]V1{{"A"}, {"B"}, {"C"}},
		[]V2{{"A", "C"}, {"B", "B"}, {"C", "A"}},
		CmpOnly("FieldOne"),
	)
	SliceCmp(
		t,
		[]V1{{"A"}, {"B"}, {"C"}},
		[]V3{{"A", "C"}, {"B", "B"}, {"C", "A"}},
		CmpRename(RenameMap{"FieldElse": "FieldOne"}),
		CmpOnly("FieldOne"),
	)
	MappedCmp(
		t,
		"FieldOne",
		"FieldTwo",
		[]V2{{"B", "B"}, {"A", "C"}, {"C", "A"}},
		[]V3{{"B", "B"}, {"A", "C"}, {"C", "A"}},
		CmpRename(RenameMap{"FieldElse": "FieldTwo"}),
		CmpOnly("FieldOne"),
	)
}

package async

import (
	"reflect"
	"testing"
)

func copyOf[T any](value T) T {
	return deepCopy(reflect.ValueOf(value)).Interface().(T)
}

func TestDeepCopyKeepsContent(t *testing.T) {
	type nested struct {
		Items  []int
		Attrs  map[string][]string
		Parent *nested
		Any    interface{}
		Fixed  [2][]int
	}
	tests := []struct {
		name  string
		value interface{}
	}{
		{name: "int", value: 7},
		{name: "string", value: "text"},
		{name: "slice", value: []int{1, 2, 3}},
		{name: "empty slice", value: []int{}},
		{name: "map", value: map[string]int{"a": 1}},
		{name: "array of slices", value: [2][]int{{1}, {2}}},
		{name: "pointer", value: &nested{Items: []int{1}}},
		{
			name: "nested struct",
			value: nested{
				Items:  []int{1},
				Attrs:  map[string][]string{"a": {"x"}},
				Parent: &nested{Any: []interface{}{1, "y"}},
				Any:    map[string]interface{}{"k": []int{2}},
				Fixed:  [2][]int{{3}, nil},
			},
		},
		{name: "nil slice", value: []int(nil)},
		{name: "nil map", value: map[string]int(nil)},
		{name: "nil pointer", value: (*nested)(nil)},
		{name: "nil interface element", value: []interface{}{nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := copyOf(tt.value)

			if !reflect.DeepEqual(got, tt.value) {
				t.Fatalf("deepCopy() = %#v, want %#v", got, tt.value)
			}
		})
	}
}

func TestDeepCopySliceHasOwnBackingArray(t *testing.T) {
	backing := []int{1, 2, 3, 4}
	original := backing[:2]

	got := copyOf(original)

	if &got[0] == &original[0] {
		t.Fatalf("deepCopy() shares the backing array")
	}
	if len(got) != 2 || cap(got) != 2 {
		t.Fatalf("deepCopy() len/cap = %d/%d, want 2/2", len(got), cap(got))
	}
}

func TestDeepCopyMapHasOwnStorage(t *testing.T) {
	original := map[string][]int{"a": {1}}

	got := copyOf(original)
	got["a"][0] = 9
	got["b"] = nil

	if original["a"][0] != 1 || len(original) != 1 {
		t.Fatalf("deepCopy() result aliases the original: %v", original)
	}
}

func TestDeepCopyPointerHasNewPointee(t *testing.T) {
	type box struct{ Items []int }
	original := &box{Items: []int{1}}

	got := copyOf(original)

	if got == original || &got.Items[0] == &original.Items[0] {
		t.Fatalf("deepCopy() shares the pointee")
	}
}

func TestDeepCopyInterfaceKeepsDynamicType(t *testing.T) {
	original := []interface{}{copyIDs{1}, map[string]int{"a": 1}}

	got := copyOf(original)

	ids, ok := got[0].(copyIDs)
	if !ok {
		t.Fatalf("deepCopy() element type = %T, want copyIDs", got[0])
	}
	if &ids[0] == &original[0].(copyIDs)[0] {
		t.Fatalf("deepCopy() shares the slice held by the interface")
	}
}

func TestDeepCopyKeepsSharedReferencesShared(t *testing.T) {
	type pair struct{ A, B *int }
	shared := 1
	original := pair{A: &shared, B: &shared}

	got := copyOf(original)

	if got.A != got.B || got.A == &shared {
		t.Fatalf("deepCopy() = %p/%p, want one new shared pointer", got.A, got.B)
	}
}

func TestDeepCopySelfReferencingMap(t *testing.T) {
	original := map[string]interface{}{}
	original["self"] = original

	got := copyOf(original)

	inner := got["self"].(map[string]interface{})
	inner["added"] = true
	if _, ok := got["added"]; !ok {
		t.Fatalf("deepCopy() did not keep the cycle in the copy")
	}
	if _, ok := original["added"]; ok {
		t.Fatalf("deepCopy() result aliases the original")
	}
}

func TestDeepCopySharesUncopyableKinds(t *testing.T) {
	type holder struct {
		Events chan int
		Run    func() int
	}
	original := holder{Events: make(chan int), Run: func() int { return 1 }}

	got := copyOf(original)

	if got.Events != original.Events || got.Run() != 1 {
		t.Fatalf("deepCopy() did not keep the channel and function")
	}
}

func TestDeepCopyUnexportedFieldIsShallow(t *testing.T) {
	type hidden struct {
		Visible []int
		hidden  []int
	}
	original := hidden{Visible: []int{1}, hidden: []int{2}}

	got := copyOf(original)

	if &got.Visible[0] == &original.Visible[0] {
		t.Fatalf("deepCopy() shares the exported field")
	}
	if &got.hidden[0] != &original.hidden[0] {
		t.Fatalf("deepCopy() copied the unexported field; documented as shared")
	}
}

func TestHasReferences(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
	}{
		{name: "int", value: 1, want: false},
		{name: "string", value: "s", want: false},
		{name: "flat struct", value: struct{ A, B int }{}, want: false},
		{name: "array of ints", value: [3]int{}, want: false},
		{name: "channel", value: make(chan int), want: false},
		{name: "slice", value: []int{}, want: true},
		{name: "pointer", value: new(int), want: true},
		{name: "map", value: map[int]int{}, want: true},
		{name: "array of interfaces", value: [1]interface{}{}, want: true},
		{name: "struct with nested slice", value: struct{ Inner struct{ S []int } }{}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasReferences(reflect.TypeOf(tt.value))

			if got != tt.want {
				t.Fatalf("hasReferences(%T) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

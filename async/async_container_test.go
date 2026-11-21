package async

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestContainerSetAndGet(t *testing.T) {
	container := &Container{}

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	container.Set("user", payload{Name: "alice", Age: 42}, nil)

	var got payload
	if err := container.Get("user", &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if got.Name != "alice" || got.Age != 42 {
		t.Fatalf("Get() = %+v, want name=alice age=42", got)
	}
}

func TestContainerGetMissingKey(t *testing.T) {
	container := &Container{}
	container.Set("present", "value", nil)

	var got string
	err := container.Get("missing", &got)
	if err == nil {
		t.Fatalf("Get() expected missing key error")
	}
}

func TestContainerGetReturnsStoredError(t *testing.T) {
	container := &Container{}
	expectedErr := errors.New("boom")
	container.Set("task", map[string]interface{}{"status": "failed"}, expectedErr)

	var got map[string]interface{}
	err := container.Get("task", &got)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("Get() error = %v, want %v", err, expectedErr)
	}

	if got["status"] != "failed" {
		t.Fatalf("Get() payload = %+v, want status=failed", got)
	}
}

func TestContainerRejectsInvalidKey(t *testing.T) {
	container := &Container{}

	defer func() {
		if recover() == nil {
			t.Fatalf("Set() expected panic for invalid key")
		}
	}()

	container.Set([]string{"invalid"}, "value", nil)
}

func TestContainerTrySetRejectsInvalidKey(t *testing.T) {
	container := &Container{}

	if err := container.TrySet([]string{"invalid"}, "value", nil); err == nil {
		t.Fatalf("TrySet() expected invalid key error")
	}
}

// unhashableKeys are keys whose static type is comparable but whose dynamic
// value is not, so hashing them into a map would panic.
var unhashableKeys = []struct {
	name string
	key  interface{}
}{
	{name: "struct with interface field holding a slice", key: struct{ X interface{} }{X: []int{1}}},
	{name: "struct with interface field holding a map", key: struct{ X interface{} }{X: map[string]int{}}},
	{name: "array of interface holding a slice", key: [2]interface{}{1, []int{1}}},
}

func TestContainerTrySetRejectsUnhashableKeyValue(t *testing.T) {
	for _, tt := range unhashableKeys {
		t.Run(tt.name, func(t *testing.T) {
			container := &Container{}

			err := container.TrySet(tt.key, "value", nil)

			if err == nil {
				t.Fatalf("TrySet() expected invalid key error")
			}
		})
	}
}

func TestContainerGetRejectsUnhashableKeyValue(t *testing.T) {
	for _, tt := range unhashableKeys {
		t.Run(tt.name, func(t *testing.T) {
			container := &Container{}
			if err := container.TrySet("other", "value", nil); err != nil {
				t.Fatalf("TrySet() error = %v", err)
			}
			var got string

			err := container.Get(tt.key, &got)

			if err == nil {
				t.Fatalf("Get() expected invalid key error")
			}
		})
	}
}

func TestContainerSetPanicsOnUnhashableKeyValue(t *testing.T) {
	container := &Container{}

	defer func() {
		if recover() == nil {
			t.Fatalf("Set() expected panic for invalid key")
		}
	}()

	container.Set(struct{ X interface{} }{X: []int{1}}, "value", nil)
}

func TestContainerAcceptsHashableInterfaceKeyValue(t *testing.T) {
	keys := []struct {
		name string
		key  interface{}
	}{
		{name: "struct with interface field holding a string", key: struct{ X interface{} }{X: "a"}},
		{name: "array of interface holding ints", key: [2]interface{}{1, 2}},
	}
	for _, tt := range keys {
		t.Run(tt.name, func(t *testing.T) {
			container := &Container{}
			var got string

			if err := container.TrySet(tt.key, "value", nil); err != nil {
				t.Fatalf("TrySet() error = %v", err)
			}
			err := container.Get(tt.key, &got)

			if err != nil || got != "value" {
				t.Fatalf("Get() = %q, %v; want %q, nil", got, err, "value")
			}
		})
	}
}

func TestContainerGetSliceIntoArray(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{name: "empty slice", length: 0, wantErr: true},
		{name: "shorter slice", length: 3, wantErr: true},
		{name: "same length", length: 4, wantErr: false},
		{name: "longer slice", length: 5, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &Container{}
			stored := make([]int, tt.length)
			for i := range stored {
				stored[i] = i + 1
			}
			container.Set("numbers", stored, nil)

			var got [4]int
			err := container.Get("numbers", &got)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Get() expected length error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if got != [4]int{1, 2, 3, 4} {
				t.Fatalf("Get() = %v, want [1 2 3 4]", got)
			}
		})
	}
}

func TestContainerGetSliceIntoArrayPointer(t *testing.T) {
	container := &Container{}
	container.Set("numbers", make([]int, 3), nil)

	var got *[4]int
	if err := container.Get("numbers", &got); err == nil {
		t.Fatalf("Get() expected length error, got %v", got)
	}
}

func FuzzContainerGetSliceIntoArray(f *testing.F) {
	f.Add(0)
	f.Add(3)
	f.Add(4)
	f.Fuzz(func(t *testing.T, length int) {
		if length < 0 || length > 8 {
			return
		}
		container := &Container{}
		container.Set("numbers", make([]int, length), nil)

		var got [4]int
		err := container.Get("numbers", &got)
		if (err != nil) != (length < len(got)) {
			t.Fatalf("Get() with length %d: error = %v", length, err)
		}
	})
}

func TestContainerGetMissingKeyQuotesKey(t *testing.T) {
	populated := &Container{}
	populated.Set("present", "value", nil)

	tests := []struct {
		name      string
		container *Container
		key       interface{}
		want      string
	}{
		{
			name:      "line break in key, empty container",
			container: &Container{},
			key:       "a\nlevel=ERROR msg=forged",
			want:      `container: item not found, key ("a\nlevel=ERROR msg=forged")`,
		},
		{
			name:      "line break in key, other keys stored",
			container: populated,
			key:       "a\nlevel=ERROR msg=forged",
			want:      `container: item not found, key ("a\nlevel=ERROR msg=forged")`,
		},
		{
			name:      "carriage return and escape sequence in key",
			container: populated,
			key:       "a\r\x1b[31mforged",
			want:      `container: item not found, key ("a\r\x1b[31mforged")`,
		},
		{
			name:      "integer key",
			container: populated,
			key:       65,
			want:      `container: item not found, key ("65")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			err := tt.container.Get(tt.key, &got)

			if err == nil || err.Error() != tt.want {
				t.Fatalf("Get() error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestContainerGetConcurrentSliceWritesDoNotRace(t *testing.T) {
	container := &Container{}
	container.Set("numbers", []int{0}, nil)

	var wg sync.WaitGroup
	for writer := 1; writer <= 2; writer++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var got []int
			if err := container.Get("numbers", &got); err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}
			got[0] = writer
		}()
	}
	wg.Wait()

	var got []int
	if err := container.Get("numbers", &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("Get() = %v, want [0]", got)
	}
}

func TestContainerGetConcurrentMapWritesDoNotRace(t *testing.T) {
	container := &Container{}
	container.Set("groups", map[string][]int{"a": {0}}, nil)

	var wg sync.WaitGroup
	for writer := 1; writer <= 2; writer++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var got map[string][]int
			if err := container.Get("groups", &got); err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}
			got["b"] = []int{writer}
			got["a"][0] = writer
		}()
	}
	wg.Wait()

	var got map[string][]int
	if err := container.Get("groups", &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if want := map[string][]int{"a": {0}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Get() = %v, want %v", got, want)
	}
}

type copyTarget struct {
	Name   string
	Items  []int
	Attrs  map[string]int
	Parent *copyTarget
}

type copyIDs []int

// getAndMutate reads key into a fresh T and mutates what it received.
func getAndMutate[T any](mutate func(got *T)) func(c *Container) error {
	return func(c *Container) error {
		var got T
		if err := c.Get("key", &got); err != nil {
			return err
		}
		mutate(&got)
		return nil
	}
}

func TestContainerGetResultIsIndependentOfStoredValue(t *testing.T) {
	tests := []struct {
		name string
		// build returns a fresh copy of the stored value on every call, so
		// one copy is stored and another is the expected content.
		build     func() interface{}
		getMutate func(c *Container) error
	}{
		{
			name:  "slice",
			build: func() interface{} { return []int{1, 2, 3} },
			getMutate: getAndMutate(func(got *[]int) {
				(*got)[0] = 9
			}),
		},
		{
			name:  "map of slices",
			build: func() interface{} { return map[string][]int{"a": {1}} },
			getMutate: getAndMutate(func(got *map[string][]int) {
				(*got)["a"][0] = 9
				(*got)["b"] = []int{2}
			}),
		},
		{
			name: "pointer to struct",
			build: func() interface{} {
				return &copyTarget{Name: "n", Items: []int{1}, Attrs: map[string]int{"a": 1}, Parent: &copyTarget{Name: "p"}}
			},
			getMutate: getAndMutate(func(got **copyTarget) {
				(*got).Name = "changed"
				(*got).Items[0] = 9
				(*got).Attrs["a"] = 9
				(*got).Parent.Name = "changed"
			}),
		},
		{
			name: "struct with reference fields",
			build: func() interface{} {
				return copyTarget{Items: []int{1}, Attrs: map[string]int{"a": 1}, Parent: &copyTarget{Name: "p"}}
			},
			getMutate: getAndMutate(func(got *copyTarget) {
				got.Items[0] = 9
				got.Attrs["a"] = 9
				got.Parent.Name = "changed"
			}),
		},
		{
			name:  "slice of interfaces into interface",
			build: func() interface{} { return []interface{}{[]int{1}, map[string]interface{}{"a": 1}} },
			getMutate: getAndMutate(func(got *interface{}) {
				values := (*got).([]interface{})
				values[0].([]int)[0] = 9
				values[1].(map[string]interface{})["a"] = 9
			}),
		},
		{
			name:  "map of interfaces into concrete type",
			build: func() interface{} { return map[string]interface{}{"list": []interface{}{1, "x"}} },
			getMutate: getAndMutate(func(got *map[string]interface{}) {
				(*got)["list"].([]interface{})[0] = 9
				(*got)["new"] = true
			}),
		},
		{
			name:  "named slice converted to slice",
			build: func() interface{} { return copyIDs{1, 2} },
			getMutate: getAndMutate(func(got *[]int) {
				(*got)[0] = 9
			}),
		},
		{
			name:  "slice converted to array pointer",
			build: func() interface{} { return []int{1, 2, 3, 4, 5} },
			getMutate: getAndMutate(func(got **[4]int) {
				(*got)[0] = 9
			}),
		},
		{
			name:  "nil references stay nil",
			build: func() interface{} { return []interface{}{nil, copyTarget{}, []int(nil), map[string]int(nil)} },
			getMutate: getAndMutate(func(got *[]interface{}) {
				(*got)[0] = 9
			}),
		},
		{
			name:  "int",
			build: func() interface{} { return 42 },
			getMutate: getAndMutate(func(got *int) {
				*got = 9
			}),
		},
		{
			name:  "string",
			build: func() interface{} { return "value" },
			getMutate: getAndMutate(func(got *string) {
				*got = "changed"
			}),
		},
		{
			name:  "flat struct",
			build: func() interface{} { return struct{ A, B int }{1, 2} },
			getMutate: getAndMutate(func(got *struct{ A, B int }) {
				got.A = 9
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &Container{}
			container.Set("key", tt.build(), nil)
			want := tt.build()

			if err := tt.getMutate(container); err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			var got interface{}
			if err := container.Get("key", &got); err != nil {
				t.Fatalf("second Get() error = %v", err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("second Get() = %#v, want %#v", got, want)
			}
		})
	}
}

type cyclicNode struct {
	Name string
	Next *cyclicNode
}

func TestContainerGetCopiesSelfReferencingPointer(t *testing.T) {
	stored := &cyclicNode{Name: "loop"}
	stored.Next = stored
	container := &Container{}
	container.Set("key", stored, nil)

	var got *cyclicNode
	err := container.Get("key", &got)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got == stored || got.Next != got || got.Name != "loop" {
		t.Fatalf("Get() = %p (next %p), stored %p; want a new self-referencing node", got, got.Next, stored)
	}
}

func TestContainerGetCopiesSelfReferencingSlice(t *testing.T) {
	stored := make([]interface{}, 1)
	stored[0] = stored
	container := &Container{}
	container.Set("key", stored, nil)

	var got []interface{}
	err := container.Get("key", &got)

	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	inner := got[0].([]interface{})
	if &got[0] == &stored[0] || &inner[0] != &got[0] {
		t.Fatalf("Get() did not return a new self-referencing slice")
	}
}

type withHidden struct {
	Name  string
	Items []int
	items []int
}

func getWithHidden(t *testing.T, stored withHidden) withHidden {
	t.Helper()
	container := &Container{}
	container.Set("key", stored, nil)

	var got withHidden
	if err := container.Get("key", &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	return got
}

func TestContainerGetStructDeepCopiesExportedFields(t *testing.T) {
	stored := withHidden{Name: "n", Items: []int{1}, items: []int{2}}

	got := getWithHidden(t, stored)

	if got.Name != "n" || &got.Items[0] == &stored.Items[0] {
		t.Fatalf("Get() = %+v; want exported fields deep-copied", got)
	}
}

func TestContainerGetStructSharesUnexportedFields(t *testing.T) {
	stored := withHidden{Name: "n", Items: []int{1}, items: []int{2}}

	got := getWithHidden(t, stored)

	// Unexported fields cannot be set through reflect, so they stay shared.
	if &got.items[0] != &stored.items[0] {
		t.Fatalf("Get() copied the unexported field; documented as shared")
	}
}

func TestContainerGetSharesPointerToStructWithUnexportedFields(t *testing.T) {
	stored := &strings.Builder{}
	stored.WriteString("a")
	container := &Container{}
	container.Set("key", stored, nil)

	var got *strings.Builder
	if err := container.Get("key", &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	got.WriteString("b")

	if got != stored {
		t.Fatalf("Get() = %p, want the stored pointer %p", got, stored)
	}
}

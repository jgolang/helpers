package async

import (
	"errors"
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

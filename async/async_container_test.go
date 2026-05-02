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

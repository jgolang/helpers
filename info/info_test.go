package info

import (
	"testing"
)

type test struct {
	Test string
	Info Info
}

func TestInfo_Set(t *testing.T) {
	info := &Info{}
	info.Set("test", "value")
	info.Set("test2", "value2")

	if got := info.GetString("test"); got != "value" {
		t.Fatalf("GetString() = %q, want %q", got, "value")
	}

	if got := info.GetString("test2"); got != "value2" {
		t.Fatalf("GetString() = %q, want %q", got, "value2")
	}
}

func TestInfo_LookupReportsPresenceAndType(t *testing.T) {
	info := Info{
		"name": "josue",
		"age":  30,
	}

	if got, ok := info.LookupString("name"); !ok || got != "josue" {
		t.Fatalf("LookupString() = (%q, %v), want (%q, true)", got, ok, "josue")
	}

	if _, ok := info.LookupString("age"); ok {
		t.Fatalf("LookupString() for wrong type should fail")
	}

	if _, ok := info.LookupString("missing"); ok {
		t.Fatalf("LookupString() for missing key should fail")
	}
}

func TestInfo_GetStruct(t *testing.T) {
	info := Info{
		"user": map[string]interface{}{
			"name": "alice",
			"age":  42,
		},
	}

	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var got user
	if err := info.GetStruct("user", &got); err != nil {
		t.Fatalf("GetStruct() error = %v", err)
	}

	if got.Name != "alice" || got.Age != 42 {
		t.Fatalf("GetStruct() = %+v, want name=alice age=42", got)
	}
}

func TestInfo_GetStructMissingKey(t *testing.T) {
	info := Info{}

	var got struct{}
	if err := info.GetStruct("missing", &got); err == nil {
		t.Fatalf("GetStruct() expected missing key error")
	}
}

func TestInfo_GetStructRejectsInvalidDestination(t *testing.T) {
	info := Info{"user": map[string]interface{}{"name": "alice"}}

	var got struct{}
	if err := info.GetStruct("user", got); err == nil {
		t.Fatalf("GetStruct() expected invalid destination error")
	}
}

func TestStructAdditonalInfo(t *testing.T) {
	var data test
	data.Info.Set("key", "value")

	if got := data.Info.GetString("key"); got != "value" {
		t.Fatalf("GetString() = %q, want %q", got, "value")
	}
}

package info

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Info stores arbitrary extension metadata for a resource.
type Info map[string]interface{}

// Set stores an additional info value.
func (data *Info) Set(key string, value interface{}) {
	if *data == nil {
		*data = Info{}
	}
	(*data)[key] = value
}

// Get returns the raw value for key.
// It returns nil when the map or key is absent.
func (data Info) Get(key string) (value interface{}) {
	if data == nil {
		return nil
	}
	return data[key]
}

// Lookup returns the raw value for key and whether the key exists.
func (data Info) Lookup(key string) (value interface{}, ok bool) {
	if data == nil {
		return nil, false
	}

	value, ok = data[key]
	return value, ok
}

// GetString returns the value as string or "" when missing or of a different type.
// Prefer LookupString when callers need to distinguish missing keys from zero values.
func (data Info) GetString(key string) string {
	value, _, ok := lookupAs[string](data, key)
	if !ok {
		return ""
	}
	return value
}

// GetInt returns the value as int or 0 when missing or of a different type.
// Prefer LookupInt when callers need to distinguish missing keys from zero values.
func (data Info) GetInt(key string) int {
	value, _, ok := lookupAs[int](data, key)
	if !ok {
		return 0
	}
	return value
}

// GetInt64 returns the value as int64 or 0 when missing or of a different type.
// Prefer LookupInt64 when callers need to distinguish missing keys from zero values.
func (data Info) GetInt64(key string) int64 {
	value, _, ok := lookupAs[int64](data, key)
	if !ok {
		return 0
	}
	return value
}

// GetFloat returns the value as float64 or 0 when missing or of a different type.
// Prefer LookupFloat when callers need to distinguish missing keys from zero values.
func (data Info) GetFloat(key string) float64 {
	value, _, ok := lookupAs[float64](data, key)
	if !ok {
		return 0.0
	}
	return value
}

// GetBool returns the value as bool or false when missing or of a different type.
// Prefer LookupBool when callers need to distinguish missing keys from zero values.
func (data Info) GetBool(key string) bool {
	value, _, ok := lookupAs[bool](data, key)
	if !ok {
		return false
	}
	return value
}

// GetStruct decodes the value stored at key into dst.
// dst must be a non-nil pointer.
func (data Info) GetStruct(key string, v interface{}) error {
	if data == nil {
		return fmt.Errorf("info: struct not found")
	}

	value, ok := data.Lookup(key)
	if !ok {
		return fmt.Errorf("info: key %q not found", key)
	}

	if err := validateDestination(v); err != nil {
		return err
	}

	if value == nil {
		return fmt.Errorf("info: key %q has nil value", key)
	}

	dstValue := reflect.ValueOf(v).Elem()
	srcValue := reflect.ValueOf(value)
	if srcValue.Type().AssignableTo(dstValue.Type()) {
		dstValue.Set(srcValue)
		return nil
	}

	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}

// LookupString gets a string value and reports whether the lookup succeeded.
func (data Info) LookupString(key string) (string, bool) {
	value, _, ok := lookupAs[string](data, key)
	return value, ok
}

// LookupInt gets an int value and reports whether the lookup succeeded.
func (data Info) LookupInt(key string) (int, bool) {
	value, _, ok := lookupAs[int](data, key)
	return value, ok
}

// LookupInt64 gets an int64 value and reports whether the lookup succeeded.
func (data Info) LookupInt64(key string) (int64, bool) {
	value, _, ok := lookupAs[int64](data, key)
	return value, ok
}

// LookupFloat gets a float64 value and reports whether the lookup succeeded.
func (data Info) LookupFloat(key string) (float64, bool) {
	value, _, ok := lookupAs[float64](data, key)
	return value, ok
}

// LookupBool gets a bool value and reports whether the lookup succeeded.
func (data Info) LookupBool(key string) (bool, bool) {
	value, _, ok := lookupAs[bool](data, key)
	return value, ok
}

func lookupAs[T any](data Info, key string) (T, bool, bool) {
	var zero T

	raw, ok := data.Lookup(key)
	if !ok {
		return zero, false, false
	}

	value, valid := raw.(T)
	if !valid {
		return zero, true, false
	}

	return value, true, true
}

func validateDestination(v interface{}) error {
	if v == nil {
		return fmt.Errorf("info: destination cannot be nil")
	}

	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("info: destination must be a non-nil pointer")
	}

	return nil
}

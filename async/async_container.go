package async

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// Container stores values and per-key errors for concurrent producers/consumers.
type Container struct {
	mu      sync.RWMutex
	content map[interface{}]interface{}
	errors  map[interface{}]error
}

// Set stores a value and optional error for a key.
// The container keeps value as given, so the caller must not modify it after
// storing it.
// It panics if the key is nil or not comparable; prefer TrySet when the key is not trusted.
func (c *Container) Set(key interface{}, value interface{}, err error) {
	if err := c.TrySet(key, value, err); err != nil {
		panic(err)
	}
}

// TrySet stores a value and optional error for a key and returns validation errors.
// The container keeps value as given, so the caller must not modify it after
// storing it.
func (c *Container) TrySet(key interface{}, value interface{}, err error) error {
	if err := validateKey(key); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.content == nil {
		c.content = map[interface{}]interface{}{}
	}

	c.content[key] = value

	if c.errors == nil {
		c.errors = map[interface{}]error{}
	}

	c.errors[key] = err

	return nil
}

// Get stores a deep copy of the value stored for key into value and returns
// any error associated with the key. value must be a non-nil pointer.
//
// The copy shares no pointed-to, slice, or map data with the stored value or
// with other callers, so callers may modify what they receive. Limits:
//   - unexported struct fields are copied shallowly, so data they reference
//     stays shared;
//   - map keys, channels, functions, and unsafe pointers are shared;
//   - when the stored value is neither assignable nor convertible to the
//     type value points to, it is decoded through a JSON round trip instead.
func (c *Container) Get(key interface{}, value interface{}) error {
	if err := validateKey(key); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	content, ok := c.content[key]
	if !ok {
		return notFoundError(key)
	}

	if err := assignValue(value, content); err != nil {
		return err
	}

	if c.errors != nil && c.errors[key] != nil {
		return c.errors[key]
	}

	return nil
}

// notFoundError quotes the key so that a key from untrusted input cannot
// inject line breaks or other control characters into logged error text.
func notFoundError(key interface{}) error {
	return fmt.Errorf("container: item not found, key (%q)", fmt.Sprint(key))
}

func validateKey(key interface{}) error {
	if key == nil {
		return fmt.Errorf("container: key cannot be nil")
	}

	// Value.Comparable also inspects dynamic values held in interface fields
	// and array elements, which Type.Comparable cannot see and which would
	// make the map insertion or lookup panic.
	if !reflect.ValueOf(key).Comparable() {
		return fmt.Errorf("container: key type %s is not comparable", reflect.TypeOf(key))
	}

	return nil
}

func assignValue(dst interface{}, src interface{}) error {
	if dst == nil {
		return fmt.Errorf("container: destination cannot be nil")
	}

	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr || dstValue.IsNil() {
		return fmt.Errorf("container: destination must be a non-nil pointer")
	}

	dstElem := dstValue.Elem()
	if src == nil {
		dstElem.Set(reflect.Zero(dstElem.Type()))
		return nil
	}

	srcValue := reflect.ValueOf(src)
	if srcValue.Type().AssignableTo(dstElem.Type()) {
		dstElem.Set(deepCopy(srcValue))
		return nil
	}

	if srcValue.Type().ConvertibleTo(dstElem.Type()) {
		if err := checkSliceToArrayLength(srcValue, dstElem.Type()); err != nil {
			return err
		}
		dstElem.Set(deepCopy(srcValue).Convert(dstElem.Type()))
		return nil
	}

	buf, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, dst)
}

// checkSliceToArrayLength reports the case where reflect's Convert would panic:
// converting a slice to an array, or to a pointer to an array, that is longer
// than the slice.
func checkSliceToArrayLength(src reflect.Value, dstType reflect.Type) error {
	if src.Kind() != reflect.Slice {
		return nil
	}

	arrayType := dstType
	if arrayType.Kind() == reflect.Ptr {
		arrayType = arrayType.Elem()
	}
	if arrayType.Kind() != reflect.Array {
		return nil
	}

	if src.Len() < arrayType.Len() {
		return fmt.Errorf("container: cannot convert slice with length %d to %s", src.Len(), dstType)
	}

	return nil
}

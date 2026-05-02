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
// It panics if the key is nil or not comparable; prefer TrySet when the key is not trusted.
func (c *Container) Set(key interface{}, value interface{}, err error) {
	if err := c.TrySet(key, value, err); err != nil {
		panic(err)
	}
}

// TrySet stores a value and optional error for a key and returns validation errors.
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

// Get copies the stored value into dst and returns any error associated with the key.
// dst must be a non-nil pointer.
func (c *Container) Get(key interface{}, value interface{}) error {
	if err := validateKey(key); err != nil {
		return err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// content:
	if c.content == nil {
		return fmt.Errorf("container: item not found, key (%v)", key)
	}

	content, ok := c.content[key]
	if !ok {
		return fmt.Errorf("container: item not found, key (%v)", key)
	}

	if err := assignValue(value, content); err != nil {
		return err
	}

	if c.errors != nil && c.errors[key] != nil {
		return c.errors[key]
	}

	return nil
}

func validateKey(key interface{}) error {
	if key == nil {
		return fmt.Errorf("container: key cannot be nil")
	}

	keyType := reflect.TypeOf(key)
	if !keyType.Comparable() {
		return fmt.Errorf("container: key type %s is not comparable", keyType)
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
		dstElem.Set(srcValue)
		return nil
	}

	if srcValue.Type().ConvertibleTo(dstElem.Type()) {
		dstElem.Set(srcValue.Convert(dstElem.Type()))
		return nil
	}

	buf, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, dst)
}

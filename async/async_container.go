package async

import (
	"encoding/json"
	"fmt"
	"sync"
)

// Container async structure.
// Use to mannipulate and share data between threads in async goroutine.
type Container struct {
	sync.Mutex
	content map[interface{}]interface{}
	errors  map[interface{}]error
}

// Set add new item to Container struct content param.
func (c *Container) Set(key interface{}, value interface{}, err error) {
	c.Lock()
	defer c.Unlock()

	if c.content == nil {
		c.content = map[interface{}]interface{}{}
	}

	c.content[key] = value

	if c.errors == nil {
		c.errors = map[interface{}]error{}
	}

	c.errors[key] = err
}

// Get return one item from Container struct content param.
func (c *Container) Get(key interface{}, value interface{}) error {
	c.Lock()
	defer c.Unlock()

	// content:
	if c.content == nil {
		return fmt.Errorf("container: item not found, key (%s)", key)
	}

	buf, err := json.Marshal(c.content[key])
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, value); err != nil {
		return err
	}

	// errors:
	if c.errors == nil {
		return fmt.Errorf("container: error item not found, key (%s)", key)
	}

	if c.errors[key] != nil {
		return c.errors[key]
	}

	return nil
}

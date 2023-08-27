package info

import (
	"encoding/json"
	"fmt"
)

// You can to use this map to extend resources struct information.
type AdditionalInfo map[string]interface{}

// Set additional info value.
func (data *AdditionalInfo) Set(key string, value interface{}) {
	if *data == nil {
		*data = AdditionalInfo{}
	}
	(*data)[key] = value
}

// Get additional info value.
func (data AdditionalInfo) Get(key string) (value interface{}) {
	if data == nil {
		return ""
	}
	return data[key]
}

// GetString gets additional info value as string.
func (data AdditionalInfo) GetString(key string) string {
	value, valid := data[key].(string)
	if !valid {
		return ""
	}
	return value
}

// GetInt gets additional info value as int.
func (data AdditionalInfo) GetInt(key string) int {
	value, valid := data[key].(int)
	if !valid {
		return 0
	}
	return value
}

// GetInt64 gets additional info value as int64.
func (data AdditionalInfo) GetInt64(key string) int64 {
	value, valid := data[key].(int64)
	if !valid {
		return 0
	}
	return value
}

// GetFloat gets additional ifno value as float64.
func (data AdditionalInfo) GetFloat(key string) float64 {
	value, valid := data[key].(float64)
	if !valid {
		return 0.0
	}
	return value
}

// GetBool gets additional info value as bool.
func (data AdditionalInfo) GetBool(key string) bool {
	value, valid := data[key].(bool)
	if !valid {
		return false
	}
	return value
}

// GetStruct unmarhal a struct in additional info map.
func (data AdditionalInfo) GetStruct(key string, v interface{}) error {
	if data == nil {
		return fmt.Errorf("Error: struct not found")
	}
	buf, err := json.Marshal(data.Get(key))
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, v)
}

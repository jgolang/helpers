package async

import "reflect"

// copier deep-copies one value. Its memo maps every pointer, map, and slice
// already copied to its copy, so shared and self-referencing data keep their
// shape in the copy and the copy terminates on cycles.
type copier struct {
	memo map[referenceKey]reflect.Value
}

// referenceKey identifies a pointer, map, or slice header. Slices also carry
// len and cap, because two slices over the same array can differ in both.
type referenceKey struct {
	typ      reflect.Type
	address  uintptr
	length   int
	capacity int
}

// deepCopy returns a copy of v that shares no pointed-to, slice, or map data
// with v. Unexported struct fields are copied shallowly, because reflect
// cannot set them without unsafe; pointers to structs with unexported fields,
// map keys, channels, functions, and unsafe pointers are shared.
func deepCopy(v reflect.Value) reflect.Value {
	if !hasReferences(v.Type()) {
		return v
	}

	c := copier{memo: map[referenceKey]reflect.Value{}}
	return c.copy(v)
}

// hasReferences reports whether a value of type t can reach data that a
// deep copy must duplicate.
func hasReferences(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Interface:
		return true
	case reflect.Array:
		return hasReferences(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if hasReferences(t.Field(i).Type) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// hasUnexportedField reports whether t is a struct with at least one
// unexported field of its own; an embedded unexported struct counts.
func hasUnexportedField(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			return true
		}
	}
	return false
}

func (c *copier) copy(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Pointer:
		return c.copyPointer(v)
	case reflect.Slice:
		return c.copySlice(v)
	case reflect.Map:
		return c.copyMap(v)
	case reflect.Array:
		return c.copyArray(v)
	case reflect.Struct:
		return c.copyStruct(v)
	case reflect.Interface:
		return c.copyInterface(v)
	default:
		return v
	}
}

// copyPointer shares a pointer to a struct with unexported fields: such types
// usually own internal state (locks, buffers, a strings.Builder's self
// pointer) that a copy of the pointee would break or duplicate.
func (c *copier) copyPointer(v reflect.Value) reflect.Value {
	if v.IsNil() || hasUnexportedField(v.Type().Elem()) {
		return v
	}

	key := referenceKey{typ: v.Type(), address: v.Pointer()}
	if copied, ok := c.memo[key]; ok {
		return copied
	}

	out := reflect.New(v.Type().Elem())
	c.memo[key] = out
	out.Elem().Set(c.copy(v.Elem()))
	return out
}

func (c *copier) copySlice(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	key := referenceKey{typ: v.Type(), address: v.Pointer(), length: v.Len(), capacity: v.Cap()}
	if copied, ok := c.memo[key]; ok {
		return copied
	}

	out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	c.memo[key] = out
	if !hasReferences(v.Type().Elem()) {
		reflect.Copy(out, v)
		return out
	}

	for i := 0; i < v.Len(); i++ {
		out.Index(i).Set(c.copy(v.Index(i)))
	}
	return out
}

func (c *copier) copyMap(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	key := referenceKey{typ: v.Type(), address: v.Pointer()}
	if copied, ok := c.memo[key]; ok {
		return copied
	}

	out := reflect.MakeMapWithSize(v.Type(), v.Len())
	c.memo[key] = out
	// Reusing one key and one value holder avoids an allocation per entry
	// that MapRange's Key and Value would make.
	entryKey := reflect.New(v.Type().Key()).Elem()
	entryValue := reflect.New(v.Type().Elem()).Elem()
	entries := v.MapRange()
	for entries.Next() {
		entryKey.SetIterKey(entries)
		entryValue.SetIterValue(entries)
		out.SetMapIndex(entryKey, c.copy(entryValue))
	}
	return out
}

func (c *copier) copyArray(v reflect.Value) reflect.Value {
	if !hasReferences(v.Type().Elem()) {
		return v
	}

	out := reflect.New(v.Type()).Elem()
	for i := 0; i < v.Len(); i++ {
		out.Index(i).Set(c.copy(v.Index(i)))
	}
	return out
}

// copyStruct sets the whole struct first so unexported fields keep their
// (shared) value, then replaces only the fields reflect allows to set.
// Values read from unexported fields are never passed to Set, which would
// panic.
func (c *copier) copyStruct(v reflect.Value) reflect.Value {
	out := reflect.New(v.Type()).Elem()
	out.Set(v)
	for i := 0; i < v.NumField(); i++ {
		if field := out.Field(i); field.CanSet() {
			field.Set(c.copy(v.Field(i)))
		}
	}
	return out
}

func (c *copier) copyInterface(v reflect.Value) reflect.Value {
	if v.IsNil() {
		return v
	}

	out := reflect.New(v.Type()).Elem()
	out.Set(c.copy(v.Elem()))
	return out
}

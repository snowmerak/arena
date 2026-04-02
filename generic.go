package arena

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// Value is a generic wrapper for any type in the arena.
type Value[T any] struct {
	arena  *Arena
	offset int
}

func (v Value[T]) Get() T {
	var zero T
	switch any(zero).(type) {
	case int8:
		return any(int8(v.arena.buffer[v.offset])).(T)
	case int16:
		return any(int16(binary.LittleEndian.Uint16(v.arena.buffer[v.offset : v.offset+2]))).(T)
	case int32:
		return any(int32(binary.LittleEndian.Uint32(v.arena.buffer[v.offset : v.offset+4]))).(T)
	case int64:
		return any(int64(binary.LittleEndian.Uint64(v.arena.buffer[v.offset : v.offset+8]))).(T)
	case uint8:
		return any(v.arena.buffer[v.offset]).(T)
	case uint16:
		return any(binary.LittleEndian.Uint16(v.arena.buffer[v.offset : v.offset+2])).(T)
	case uint32:
		return any(binary.LittleEndian.Uint32(v.arena.buffer[v.offset : v.offset+4])).(T)
	case uint64:
		return any(binary.LittleEndian.Uint64(v.arena.buffer[v.offset : v.offset+8])).(T)
	case float32:
		return any(math.Float32frombits(binary.LittleEndian.Uint32(v.arena.buffer[v.offset : v.offset+4]))).(T)
	case float64:
		return any(math.Float64frombits(binary.LittleEndian.Uint64(v.arena.buffer[v.offset : v.offset+8]))).(T)
	case complex64:
		r := math.Float32frombits(binary.LittleEndian.Uint32(v.arena.buffer[v.offset : v.offset+4]))
		i := math.Float32frombits(binary.LittleEndian.Uint32(v.arena.buffer[v.offset+4 : v.offset+8]))
		return any(complex(r, i)).(T)
	case complex128:
		r := math.Float64frombits(binary.LittleEndian.Uint64(v.arena.buffer[v.offset : v.offset+8]))
		i := math.Float64frombits(binary.LittleEndian.Uint64(v.arena.buffer[v.offset+8 : v.offset+16]))
		return any(complex(r, i)).(T)
	case string:
		dataOffset := int(binary.LittleEndian.Uint64(v.arena.buffer[v.offset : v.offset+8]))
		length := int(binary.LittleEndian.Uint64(v.arena.buffer[v.offset+8 : v.offset+16]))
		return any(string(v.arena.buffer[dataOffset : dataOffset+length])).(T)
	}
	return zero
}

func (v Value[T]) Set(val T) error {
	switch any(val).(type) {
	case int8:
		v.arena.buffer[v.offset] = byte(any(val).(int8))
	case int16:
		binary.LittleEndian.PutUint16(v.arena.buffer[v.offset:v.offset+2], uint16(any(val).(int16)))
	case int32:
		binary.LittleEndian.PutUint32(v.arena.buffer[v.offset:v.offset+4], uint32(any(val).(int32)))
	case int64:
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset:v.offset+8], uint64(any(val).(int64)))
	case uint8:
		v.arena.buffer[v.offset] = any(val).(uint8)
	case uint16:
		binary.LittleEndian.PutUint16(v.arena.buffer[v.offset:v.offset+2], any(val).(uint16))
	case uint32:
		binary.LittleEndian.PutUint32(v.arena.buffer[v.offset:v.offset+4], any(val).(uint32))
	case uint64:
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset:v.offset+8], any(val).(uint64))
	case float32:
		binary.LittleEndian.PutUint32(v.arena.buffer[v.offset:v.offset+4], math.Float32bits(any(val).(float32)))
	case float64:
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset:v.offset+8], math.Float64bits(any(val).(float64)))
	case complex64:
		c := any(val).(complex64)
		binary.LittleEndian.PutUint32(v.arena.buffer[v.offset:v.offset+4], math.Float32bits(real(c)))
		binary.LittleEndian.PutUint32(v.arena.buffer[v.offset+4:v.offset+8], math.Float32bits(imag(c)))
	case complex128:
		c := any(val).(complex128)
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset:v.offset+8], math.Float64bits(real(c)))
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset+8:v.offset+16], math.Float64bits(imag(c)))
	case string:
		s := any(val).(string)
		dataOffset, err := v.arena.Alloc(len(s))
		if err != nil {
			return err
		}
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset:v.offset+8], uint64(dataOffset))
		binary.LittleEndian.PutUint64(v.arena.buffer[v.offset+8:v.offset+16], uint64(len(s)))
		copy(v.arena.buffer[dataOffset:dataOffset+len(s)], s)
	}
	return nil
}

// Array is a generic wrapper for an array of any type in the arena.
type Array[T any] struct {
	arena      *Arena
	offset     int
	length     int
	elemStride int
}

func (a Array[T]) Get(i int) (Value[T], error) {
	if i < 0 || i >= a.length {
		return Value[T]{}, fmt.Errorf("index out of bounds: %d", i)
	}
	return Value[T]{arena: a.arena, offset: a.offset + i*a.elemStride}, nil
}

func (a Array[T]) Len() int {
	return a.length
}

// Alloc is a global generic function to allocate a single value of type T.
func Alloc[T any](a *Arena) (Value[T], error) {
	var zero T
	t := reflect.TypeOf(zero)
	var size int
	if t.Kind() == reflect.Struct {
		schema, err := GetSchema(t)
		if err != nil {
			return Value[T]{}, err
		}
		size = schema.Size
	} else {
		var err error
		size, err = getKindSize(t.Kind())
		if err != nil {
			return Value[T]{}, err
		}
	}

	offset, err := a.Alloc(size)
	if err != nil {
		return Value[T]{}, err
	}

	return Value[T]{arena: a, offset: offset}, nil
}

// AllocArray is a global generic function to allocate an array of type T.
func AllocArray[T any](a *Arena, length int) (Array[T], error) {
	var zero T
	t := reflect.TypeOf(zero)
	var stride int
	if t.Kind() == reflect.Struct {
		schema, err := GetSchema(t)
		if err != nil {
			return Array[T]{}, err
		}
		stride = schema.Size
	} else {
		var err error
		stride, err = getKindSize(t.Kind())
		if err != nil {
			return Array[T]{}, err
		}
	}

	offset, err := a.Alloc(length * stride)
	if err != nil {
		return Array[T]{}, err
	}

	return Array[T]{arena: a, offset: offset, length: length, elemStride: stride}, nil
}

// GetField is a global generic function to access a field of a struct Value.
func GetField[T any, F any](v Value[T], name string) (Value[F], error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() != reflect.Struct {
		return Value[F]{}, fmt.Errorf("type %s is not a struct", t.Name())
	}
	schema, err := GetSchema(t)
	if err != nil {
		return Value[F]{}, err
	}
	info, ok := schema.Fields[name]
	if !ok {
		return Value[F]{}, fmt.Errorf("field %s not found in %s", name, t.Name())
	}

	// Basic type check
	var zeroF F
	fType := reflect.TypeOf(zeroF)
	if info.Kind != fType.Kind() {
		return Value[F]{}, fmt.Errorf("field %s kind mismatch: expected %s, got %s", name, fType.Kind(), info.Kind)
	}

	return Value[F]{arena: v.arena, offset: v.offset + info.Offset}, nil
}

// GetArrayField is a global generic function to access an array field of a struct Value.
func GetArrayField[T any, E any](v Value[T], name string) (Array[E], error) {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() != reflect.Struct {
		return Array[E]{}, fmt.Errorf("type %s is not a struct", t.Name())
	}
	schema, err := GetSchema(t)
	if err != nil {
		return Array[E]{}, err
	}
	info, ok := schema.Fields[name]
	if !ok {
		return Array[E]{}, fmt.Errorf("field %s not found in %s", name, t.Name())
	}

	if info.Kind != reflect.Array {
		return Array[E]{}, fmt.Errorf("field %s is not an array", name)
	}

	var zeroE E
	eType := reflect.TypeOf(zeroE)
	if info.ElemKind != eType.Kind() {
		return Array[E]{}, fmt.Errorf("field %s element kind mismatch: expected %s, got %s", name, eType.Kind(), info.ElemKind)
	}

	var stride int
	if info.ElemKind == reflect.Struct {
		stride = info.ElemSchema.Size
	} else {
		var err error
		stride, err = getKindSize(info.ElemKind)
		if err != nil {
			return Array[E]{}, err
		}
	}

	return Array[E]{arena: v.arena, offset: v.offset + info.Offset, length: info.Len, elemStride: stride}, nil
}

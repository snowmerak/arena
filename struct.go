package arena

import (
	"fmt"
	"reflect"
	"sync"
)

var schemaCache sync.Map // map[reflect.Type]*StructSchema

// FieldInfo contains offset and type information for a mapped struct field.
type FieldInfo struct {
	Offset int
	Kind   reflect.Kind
}

// StructSchema stores the total packed size and field mappings for a struct.
type StructSchema struct {
	Size   int
	Fields map[string]FieldInfo
}

// buildSchema uses reflection to tightly pack primitive fields sequentially.
func buildSchema(t reflect.Type) (*StructSchema, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", t.Kind())
	}

	schema := &StructSchema{
		Fields: make(map[string]FieldInfo),
	}

	offset := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		kind := f.Type.Kind()

		var size int
		switch kind {
		case reflect.Int8, reflect.Uint8:
			size = 1
		case reflect.Int16, reflect.Uint16:
			size = 2
		case reflect.Int32, reflect.Uint32, reflect.Float32:
			size = 4
		case reflect.Int64, reflect.Uint64, reflect.Float64:
			size = 8
		default:
			return nil, fmt.Errorf("unsupported field type %s for field %s", kind, f.Name)
		}

		schema.Fields[f.Name] = FieldInfo{Offset: offset, Kind: kind}
		offset += size
	}

	schema.Size = offset
	return schema, nil
}

// GetSchema returns a cached struct schema, evaluating it if it's the first time.
func GetSchema(t reflect.Type) (*StructSchema, error) {
	if val, ok := schemaCache.Load(t); ok {
		return val.(*StructSchema), nil
	}

	schema, err := buildSchema(t)
	if err != nil {
		return nil, err
	}

	schemaCache.Store(t, schema)
	return schema, nil
}

// StructInstance wraps a specific memory chunk in the Arena governed by a struct schema.
type StructInstance struct {
	arena  *Arena
	offset int
	schema *StructSchema
}

// AllocStruct calculates and reserves the tight-packed size required for the struct type 'T'.
// Note: Due to lack of type parameters in older Go versions or to keep API broad, we accept any interface{}.
func (a *Arena) AllocStruct(target interface{}) (StructInstance, error) {
	t := reflect.TypeOf(target)

	schema, err := GetSchema(t)
	if err != nil {
		return StructInstance{}, err
	}

	offset, err := a.Alloc(schema.Size)
	if err != nil {
		return StructInstance{}, err
	}

	return StructInstance{arena: a, offset: offset, schema: schema}, nil
}

// -- Typed Wrappers Generation --

func (s StructInstance) FieldInt8(name string) (Int8, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Int8 {
		return Int8{}, fmt.Errorf("field %s missing or not int8", name)
	}
	return Int8{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldInt16(name string) (Int16, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Int16 {
		return Int16{}, fmt.Errorf("field %s missing or not int16", name)
	}
	return Int16{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldInt32(name string) (Int32, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Int32 {
		return Int32{}, fmt.Errorf("field %s missing or not int32", name)
	}
	return Int32{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldInt64(name string) (Int64, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Int64 {
		return Int64{}, fmt.Errorf("field %s missing or not int64", name)
	}
	return Int64{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldUint8(name string) (Uint8, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Uint8 {
		return Uint8{}, fmt.Errorf("field %s missing or not uint8", name)
	}
	return Uint8{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldUint16(name string) (Uint16, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Uint16 {
		return Uint16{}, fmt.Errorf("field %s missing or not uint16", name)
	}
	return Uint16{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldUint32(name string) (Uint32, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Uint32 {
		return Uint32{}, fmt.Errorf("field %s missing or not uint32", name)
	}
	return Uint32{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldUint64(name string) (Uint64, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Uint64 {
		return Uint64{}, fmt.Errorf("field %s missing or not uint64", name)
	}
	return Uint64{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldFloat32(name string) (Float32, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Float32 {
		return Float32{}, fmt.Errorf("field %s missing or not float32", name)
	}
	return Float32{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldFloat64(name string) (Float64, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Float64 {
		return Float64{}, fmt.Errorf("field %s missing or not float64", name)
	}
	return Float64{arena: s.arena, offset: s.offset + info.Offset}, nil
}

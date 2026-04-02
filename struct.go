package arena

import (
	"fmt"
	"reflect"
	"sync"
)

var schemaCache sync.Map // map[reflect.Type]*StructSchema

// FieldInfo contains offset and type information for a mapped struct field.
type FieldInfo struct {
	Offset     int
	Kind       reflect.Kind
	ElemKind   reflect.Kind  // For arrays
	Len        int           // For arrays
	ElemSchema *StructSchema // For arrays of structs
}

// StructSchema stores the total packed size and field mappings for a struct.
type StructSchema struct {
	Size   int
	Fields map[string]FieldInfo
}

func getKindSize(kind reflect.Kind) (int, error) {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 1, nil
	case reflect.Int16, reflect.Uint16:
		return 2, nil
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 4, nil
	case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Complex64:
		return 8, nil
	case reflect.Complex128, reflect.String:
		return 16, nil
	default:
		return 0, fmt.Errorf("unsupported kind %s", kind)
	}
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
		var elemKind reflect.Kind
		var arrayLen int
		var elemSchema *StructSchema

		if kind == reflect.Array {
			elemType := f.Type.Elem()
			elemKind = elemType.Kind()
			var elemSize int
			if elemKind == reflect.Struct {
				var err error
				elemSchema, err = GetSchema(elemType)
				if err != nil {
					return nil, err
				}
				elemSize = elemSchema.Size
			} else {
				var err error
				elemSize, err = getKindSize(elemKind)
				if err != nil {
					return nil, fmt.Errorf("unsupported array element type %s for field %s", elemKind, f.Name)
				}
			}
			arrayLen = f.Type.Len()
			size = elemSize * arrayLen
		} else if kind == reflect.Struct {
			var err error
			elemSchema, err = GetSchema(f.Type)
			if err != nil {
				return nil, err
			}
			size = elemSchema.Size
		} else {
			var err error
			size, err = getKindSize(kind)
			if err != nil {
				return nil, fmt.Errorf("unsupported field type %s for field %s", kind, f.Name)
			}
		}

		schema.Fields[f.Name] = FieldInfo{
			Offset:     offset,
			Kind:       kind,
			ElemKind:   elemKind,
			Len:        arrayLen,
			ElemSchema: elemSchema,
		}
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

func (s StructInstance) FieldComplex64(name string) (Complex64, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Complex64 {
		return Complex64{}, fmt.Errorf("field %s missing or not complex64", name)
	}
	return Complex64{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldComplex128(name string) (Complex128, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Complex128 {
		return Complex128{}, fmt.Errorf("field %s missing or not complex128", name)
	}
	return Complex128{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldString(name string) (String, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.String {
		return String{}, fmt.Errorf("field %s missing or not string", name)
	}
	return String{arena: s.arena, offset: s.offset + info.Offset}, nil
}

func (s StructInstance) FieldArray(name string) (any, error) {
	info, ok := s.schema.Fields[name]
	if !ok || info.Kind != reflect.Array {
		return nil, fmt.Errorf("field %s missing or not array", name)
	}

	if info.ElemKind == reflect.Struct {
		return StructArrayInstance{
			arena:      s.arena,
			offset:     s.offset + info.Offset,
			length:     info.Len,
			elemStride: info.ElemSchema.Size,
			schema:     info.ElemSchema,
		}, nil
	}

	elemStride, err := getKindSize(info.ElemKind)
	if err != nil {
		return nil, err
	}

	switch info.ElemKind {
	case reflect.Int8:
		return ArrayInstance[Int8]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Int16:
		return ArrayInstance[Int16]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Int32:
		return ArrayInstance[Int32]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Int64:
		return ArrayInstance[Int64]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Uint8:
		return ArrayInstance[Uint8]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Uint16:
		return ArrayInstance[Uint16]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Uint32:
		return ArrayInstance[Uint32]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Uint64:
		return ArrayInstance[Uint64]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Float32:
		return ArrayInstance[Float32]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Float64:
		return ArrayInstance[Float64]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Complex64:
		return ArrayInstance[Complex64]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.Complex128:
		return ArrayInstance[Complex128]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	case reflect.String:
		return ArrayInstance[String]{arena: s.arena, offset: s.offset + info.Offset, length: info.Len, elemStride: elemStride}, nil
	default:
		return nil, fmt.Errorf("unsupported array element type %s", info.ElemKind)
	}
}

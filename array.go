package arena

import (
	"fmt"
	"reflect"
)

// ArrayInstance represents a fixed-size array in the arena.
type ArrayInstance[T Primitive] struct {
	arena      *Arena
	offset     int
	length     int
	elemStride int
}

func (a ArrayInstance[T]) at(i int) (int, error) {
	if i < 0 || i >= a.length {
		return 0, fmt.Errorf("index out of bounds: %d", i)
	}
	return a.offset + i*a.elemStride, nil
}

func (a ArrayInstance[T]) Get(i int) (T, error) {
	offset, err := a.at(i)
	if err != nil {
		var zero T
		return zero, err
	}
	return T{arena: a.arena, offset: offset}, nil
}

func (a *Arena) AllocArray(target any) (any, error) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected array type, got %s", t.Kind())
	}

	elemKind := t.Elem().Kind()
	stride, err := getKindSize(elemKind)
	if err != nil {
		return nil, err
	}

	length := t.Len()
	offset, err := a.Alloc(length * stride)
	if err != nil {
		return nil, err
	}

	switch elemKind {
	case reflect.Int8:
		return ArrayInstance[Int8]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Int16:
		return ArrayInstance[Int16]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Int32:
		return ArrayInstance[Int32]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Int64:
		return ArrayInstance[Int64]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Uint8:
		return ArrayInstance[Uint8]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Uint16:
		return ArrayInstance[Uint16]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Uint32:
		return ArrayInstance[Uint32]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Uint64:
		return ArrayInstance[Uint64]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Float32:
		return ArrayInstance[Float32]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Float64:
		return ArrayInstance[Float64]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Complex64:
		return ArrayInstance[Complex64]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.Complex128:
		return ArrayInstance[Complex128]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	case reflect.String:
		return ArrayInstance[String]{arena: a, offset: offset, length: length, elemStride: stride}, nil
	default:
		return nil, fmt.Errorf("unsupported array element type %s", elemKind)
	}
}

// StructArrayInstance represents a fixed-size array of structs in the arena.
type StructArrayInstance struct {
	arena      *Arena
	offset     int
	length     int
	elemStride int
	schema     *StructSchema
}

func (a StructArrayInstance) at(i int) (int, error) {
	if i < 0 || i >= a.length {
		return 0, fmt.Errorf("index out of bounds: %d", i)
	}
	return a.offset + i*a.elemStride, nil
}

func (a StructArrayInstance) Get(i int) (StructInstance, error) {
	offset, err := a.at(i)
	if err != nil {
		return StructInstance{}, err
	}
	return StructInstance{arena: a.arena, offset: offset, schema: a.schema}, nil
}

func (a *Arena) AllocStructArray(target any) (StructArrayInstance, error) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Array {
		return StructArrayInstance{}, fmt.Errorf("expected array type, got %s", t.Kind())
	}

	elemType := t.Elem()
	if elemType.Kind() != reflect.Struct {
		return StructArrayInstance{}, fmt.Errorf("expected array of structs, got array of %s", elemType.Kind())
	}

	schema, err := GetSchema(elemType)
	if err != nil {
		return StructArrayInstance{}, err
	}

	length := t.Len()
	offset, err := a.Alloc(length * schema.Size)
	if err != nil {
		return StructArrayInstance{}, err
	}

	return StructArrayInstance{
		arena:      a,
		offset:     offset,
		length:     length,
		elemStride: schema.Size,
		schema:     schema,
	}, nil
}

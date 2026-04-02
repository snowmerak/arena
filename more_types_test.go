package arena_test

import (
	"testing"
	"github.com/snowmerak/arena"
)

func TestArena_MorePrimitives(t *testing.T) {
	a := arena.New(1024)

	c64, err := a.AllocComplex64()
	if err != nil {
		t.Fatalf("failed to alloc complex64: %v", err)
	}
	c64.Set(complex(1.2, 3.4))
	if v := c64.Get(); v != complex(float32(1.2), float32(3.4)) {
		t.Errorf("expected (1.2+3.4i), got %v", v)
	}

	c128, err := a.AllocComplex128()
	if err != nil {
		t.Fatalf("failed to alloc complex128: %v", err)
	}
	c128.Set(complex(5.6, 7.8))
	if v := c128.Get(); v != complex(5.6, 7.8) {
		t.Errorf("expected (5.6+7.8i), got %v", v)
	}

	s, err := a.AllocString("hello arena")
	if err != nil {
		t.Fatalf("failed to alloc string: %v", err)
	}
	if v := s.Get(); v != "hello arena" {
		t.Errorf("expected 'hello arena', got '%s'", v)
	}

	err = s.Set("new string")
	if err != nil {
		t.Fatalf("failed to set string: %v", err)
	}
	if v := s.Get(); v != "new string" {
		t.Errorf("expected 'new string', got '%s'", v)
	}
}

func TestArena_Array(t *testing.T) {
	a := arena.New(1024)

	rawArr, err := a.AllocArray([5]int32{})
	if err != nil {
		t.Fatalf("failed to alloc array: %v", err)
	}
	arr := rawArr.(arena.ArrayInstance[arena.Int32])

	for i := 0; i < 5; i++ {
		el, _ := arr.Get(i)
		el.Set(int32(i * 10))
	}

	for i := 0; i < 5; i++ {
		el, _ := arr.Get(i)
		if v := el.Get(); v != int32(i*10) {
			t.Errorf("at index %d: expected %d, got %d", i, i*10, v)
		}
	}
}

type SubStruct struct {
	ID    int32
	Value float32
}

type ParentStruct struct {
	Items [3]SubStruct
}

func TestArena_StructArray(t *testing.T) {
	a := arena.New(1024)

	var dummy ParentStruct
	inst, err := a.AllocStruct(dummy)
	if err != nil {
		t.Fatalf("failed to alloc struct: %v", err)
	}

	rawArr, err := inst.FieldArray("Items")
	if err != nil {
		t.Fatalf("failed to get Items field: %v", err)
	}
	arr := rawArr.(arena.StructArrayInstance)

	for i := 0; i < 3; i++ {
		sub, _ := arr.Get(i)
		id, _ := sub.FieldInt32("ID")
		val, _ := sub.FieldFloat32("Value")
		id.Set(int32(i + 1))
		val.Set(float32(i) * 1.5)
	}

	for i := 0; i < 3; i++ {
		sub, _ := arr.Get(i)
		id, _ := sub.FieldInt32("ID")
		val, _ := sub.FieldFloat32("Value")
		if v := id.Get(); v != int32(i+1) {
			t.Errorf("at index %d: expected ID %d, got %d", i, i+1, v)
		}
		if v := val.Get(); v != float32(i)*1.5 {
			t.Errorf("at index %d: expected Value %f, got %f", i, float32(i)*1.5, v)
		}
	}
}

type ExtendedStruct struct {
	C64  complex64
	C128 complex128
	S    string
	Arr  [3]float32
}

func TestArena_ExtendedStruct(t *testing.T) {
	a := arena.New(2048)

	var dummy ExtendedStruct
	inst, err := a.AllocStruct(dummy)
	if err != nil {
		t.Fatalf("failed to alloc struct: %v", err)
	}

	c64, _ := inst.FieldComplex64("C64")
	c64.Set(complex(1, 2))

	c128, _ := inst.FieldComplex128("C128")
	c128.Set(complex(3, 4))

	s, _ := inst.FieldString("S")
	s.Set("struct string")

	rawArr, _ := inst.FieldArray("Arr")
	arr := rawArr.(arena.ArrayInstance[arena.Float32])
	for i := 0; i < 3; i++ {
		el, _ := arr.Get(i)
		el.Set(float32(i) + 0.5)
	}

	// Verify
	if v := c64.Get(); v != complex(float32(1), float32(2)) {
		t.Errorf("C64: expected (1+2i), got %v", v)
	}
	if v := c128.Get(); v != complex(3, 4) {
		t.Errorf("C128: expected (3+4i), got %v", v)
	}
	if v := s.Get(); v != "struct string" {
		t.Errorf("S: expected 'struct string', got '%s'", v)
	}
	for i := 0; i < 3; i++ {
		el, _ := arr.Get(i)
		if v := el.Get(); v != float32(i)+0.5 {
			t.Errorf("Arr[%d]: expected %f, got %f", i, float32(i)+0.5, v)
		}
	}
}

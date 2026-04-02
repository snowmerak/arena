package arena_test

import (
	"testing"

	"github.com/snowmerak/arena"
)

func TestArena_Primitives(t *testing.T) {
	a := arena.New(1024)

	i32, err := a.AllocInt32()
	if err != nil {
		t.Fatalf("failed to alloc int32: %v", err)
	}

	i32.Set(123456789)
	if v := i32.Get(); v != 123456789 {
		t.Errorf("expected 123456789, got %d", v)
	}

	f64, err := a.AllocFloat64()
	if err != nil {
		t.Fatalf("failed to alloc float64: %v", err)
	}

	f64.Set(3.14159)
	if v := f64.Get(); v != 3.14159 {
		t.Errorf("expected 3.14159, got %f", v)
	}

	// Verify offset increment correctly (4 bytes for int32 + 8 bytes for float64 = 12 bytes total)
	u8, err := a.AllocUint8()
	if err != nil {
		t.Fatalf("failed to alloc uint8: %v", err)
	}
	
	// Since uint8 is the 3rd allocation, its offset should be 12 (0 + 4 + 8)
	// We can't access 'offset' field directly as it's private, but we can verify buffer size
	
	u8.Set(255)
	if v := u8.Get(); v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
}

func TestArena_MapRegistry(t *testing.T) {
	a := arena.New(512)

	offset, err := a.Alloc(100)
	if err != nil {
		t.Fatal(err)
	}

	a.SetName("block1", offset)

	retrieved, ok := a.GetName("block1")
	if !ok || retrieved != offset {
		t.Errorf("expected %d, got %d", offset, retrieved)
	}

	a.Reset()
	_, ok = a.GetName("block1")
	if ok {
		t.Errorf("expected registry to be cleared on Reset")
	}
}

type TestStruct struct {
	Count int32
	Value float64
	Flag  uint8
}

func TestArena_StructMapping(t *testing.T) {
	a := arena.New(1024)

	var dummy TestStruct
	inst, err := a.AllocStruct(dummy)
	if err != nil {
		t.Fatalf("failed to alloc struct: %v", err)
	}

	c, err := inst.FieldInt32("Count")
	if err != nil {
		t.Fatalf("failed to get count wrapper: %v", err)
	}
	c.Set(42)

	v, err := inst.FieldFloat64("Value")
	if err != nil {
		t.Fatalf("failed to get value wrapper: %v", err)
	}
	v.Set(3.14)

	f, err := inst.FieldUint8("Flag")
	if err != nil {
		t.Fatalf("failed to get flag wrapper: %v", err)
	}
	f.Set(1)

	// Verify retrieval
	cc, _ := inst.FieldInt32("Count")
	if val := cc.Get(); val != 42 {
		t.Errorf("expected Count 42, got %d", val)
	}

	vv, _ := inst.FieldFloat64("Value")
	if val := vv.Get(); val != 3.14 {
		t.Errorf("expected Value 3.14, got %f", val)
	}

	ff, _ := inst.FieldUint8("Flag")
	if val := ff.Get(); val != 1 {
		t.Errorf("expected Flag 1, got %d", val)
	}
}

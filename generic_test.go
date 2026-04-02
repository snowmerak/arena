package arena_test

import (
	"testing"
	"github.com/snowmerak/arena"
)

func TestArena_GenericAlloc(t *testing.T) {
	a := arena.New(1024)

	i32, err := arena.Alloc[int32](a)
	if err != nil {
		t.Fatal(err)
	}
	i32.Set(100)
	if v := i32.Get(); v != 100 {
		t.Errorf("expected 100, got %d", v)
	}

	str, err := arena.Alloc[string](a)
	if err != nil {
		t.Fatal(err)
	}
	str.Set("generic string")
	if v := str.Get(); v != "generic string" {
		t.Errorf("expected 'generic string', got '%s'", v)
	}
}

func TestArena_GenericArray(t *testing.T) {
	a := arena.New(1024)

	arr, err := arena.AllocArray[float64](a, 3)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < arr.Len(); i++ {
		val, _ := arr.Get(i)
		val.Set(float64(i) * 1.1)
	}

	for i := 0; i < arr.Len(); i++ {
		val, _ := arr.Get(i)
		if v := val.Get(); v != float64(i)*1.1 {
			t.Errorf("at index %d: expected %f, got %f", i, float64(i)*1.1, v)
		}
	}
}

type Point struct {
	X int32
	Y int32
}

type Rectangle struct {
	TopLeft Point
	BottomRight Point
	Labels [2]string
}

func TestArena_GenericStruct(t *testing.T) {
	a := arena.New(2048)

	rect, err := arena.Alloc[Rectangle](a)
	if err != nil {
		t.Fatal(err)
	}

	// Access nested struct field
	tl, _ := arena.GetField[Rectangle, Point](rect, "TopLeft")
	tlX, _ := arena.GetField[Point, int32](tl, "X")
	tlX.Set(10)

	// Access array field
	labels, _ := arena.GetArrayField[Rectangle, string](rect, "Labels")
	l0, _ := labels.Get(0)
	l0.Set("rect-label")

	// Verify
	if v := tlX.Get(); v != 10 {
		t.Errorf("tlX: expected 10, got %d", v)
	}
	if v := l0.Get(); v != "rect-label" {
		t.Errorf("l0: expected 'rect-label', got '%s'", v)
	}
}

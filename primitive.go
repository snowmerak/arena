package arena

import (
	"encoding/binary"
	"math"
)

// Int8 wrapper
type Int8 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocInt8() (Int8, error) {
	offset, err := a.Alloc(1)
	if err != nil {
		return Int8{}, err
	}
	return Int8{arena: a, offset: offset}, nil
}

func (i Int8) Get() int8 {
	return int8(i.arena.buffer[i.offset])
}

func (i Int8) Set(v int8) {
	i.arena.buffer[i.offset] = byte(v)
}

// Int16 wrapper
type Int16 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocInt16() (Int16, error) {
	offset, err := a.Alloc(2)
	if err != nil {
		return Int16{}, err
	}
	return Int16{arena: a, offset: offset}, nil
}

func (i Int16) Get() int16 {
	return int16(binary.LittleEndian.Uint16(i.arena.buffer[i.offset : i.offset+2]))
}

func (i Int16) Set(v int16) {
	binary.LittleEndian.PutUint16(i.arena.buffer[i.offset:i.offset+2], uint16(v))
}

// Int32 wrapper
type Int32 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocInt32() (Int32, error) {
	offset, err := a.Alloc(4)
	if err != nil {
		return Int32{}, err
	}
	return Int32{arena: a, offset: offset}, nil
}

func (i Int32) Get() int32 {
	return int32(binary.LittleEndian.Uint32(i.arena.buffer[i.offset : i.offset+4]))
}

func (i Int32) Set(v int32) {
	binary.LittleEndian.PutUint32(i.arena.buffer[i.offset:i.offset+4], uint32(v))
}

// Int64 wrapper
type Int64 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocInt64() (Int64, error) {
	offset, err := a.Alloc(8)
	if err != nil {
		return Int64{}, err
	}
	return Int64{arena: a, offset: offset}, nil
}

func (i Int64) Get() int64 {
	return int64(binary.LittleEndian.Uint64(i.arena.buffer[i.offset : i.offset+8]))
}

func (i Int64) Set(v int64) {
	binary.LittleEndian.PutUint64(i.arena.buffer[i.offset:i.offset+8], uint64(v))
}

// Uint8 wrapper
type Uint8 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocUint8() (Uint8, error) {
	offset, err := a.Alloc(1)
	if err != nil {
		return Uint8{}, err
	}
	return Uint8{arena: a, offset: offset}, nil
}

func (u Uint8) Get() uint8 {
	return u.arena.buffer[u.offset]
}

func (u Uint8) Set(v uint8) {
	u.arena.buffer[u.offset] = v
}

// Uint16 wrapper
type Uint16 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocUint16() (Uint16, error) {
	offset, err := a.Alloc(2)
	if err != nil {
		return Uint16{}, err
	}
	return Uint16{arena: a, offset: offset}, nil
}

func (u Uint16) Get() uint16 {
	return binary.LittleEndian.Uint16(u.arena.buffer[u.offset : u.offset+2])
}

func (u Uint16) Set(v uint16) {
	binary.LittleEndian.PutUint16(u.arena.buffer[u.offset:u.offset+2], v)
}

// Uint32 wrapper
type Uint32 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocUint32() (Uint32, error) {
	offset, err := a.Alloc(4)
	if err != nil {
		return Uint32{}, err
	}
	return Uint32{arena: a, offset: offset}, nil
}

func (u Uint32) Get() uint32 {
	return binary.LittleEndian.Uint32(u.arena.buffer[u.offset : u.offset+4])
}

func (u Uint32) Set(v uint32) {
	binary.LittleEndian.PutUint32(u.arena.buffer[u.offset:u.offset+4], v)
}

// Uint64 wrapper
type Uint64 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocUint64() (Uint64, error) {
	offset, err := a.Alloc(8)
	if err != nil {
		return Uint64{}, err
	}
	return Uint64{arena: a, offset: offset}, nil
}

func (u Uint64) Get() uint64 {
	return binary.LittleEndian.Uint64(u.arena.buffer[u.offset : u.offset+8])
}

func (u Uint64) Set(v uint64) {
	binary.LittleEndian.PutUint64(u.arena.buffer[u.offset:u.offset+8], v)
}

// Float32 wrapper
type Float32 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocFloat32() (Float32, error) {
	offset, err := a.Alloc(4)
	if err != nil {
		return Float32{}, err
	}
	return Float32{arena: a, offset: offset}, nil
}

func (f Float32) Get() float32 {
	bits := binary.LittleEndian.Uint32(f.arena.buffer[f.offset : f.offset+4])
	return math.Float32frombits(bits)
}

func (f Float32) Set(v float32) {
	bits := math.Float32bits(v)
	binary.LittleEndian.PutUint32(f.arena.buffer[f.offset:f.offset+4], bits)
}

// Float64 wrapper
type Float64 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocFloat64() (Float64, error) {
	offset, err := a.Alloc(8)
	if err != nil {
		return Float64{}, err
	}
	return Float64{arena: a, offset: offset}, nil
}

func (f Float64) Get() float64 {
	bits := binary.LittleEndian.Uint64(f.arena.buffer[f.offset : f.offset+8])
	return math.Float64frombits(bits)
}

func (f Float64) Set(v float64) {
	bits := math.Float64bits(v)
	binary.LittleEndian.PutUint64(f.arena.buffer[f.offset:f.offset+8], bits)
}

// Complex64 wrapper
type Complex64 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocComplex64() (Complex64, error) {
	offset, err := a.Alloc(8)
	if err != nil {
		return Complex64{}, err
	}
	return Complex64{arena: a, offset: offset}, nil
}

func (c Complex64) Get() complex64 {
	r := math.Float32frombits(binary.LittleEndian.Uint32(c.arena.buffer[c.offset : c.offset+4]))
	i := math.Float32frombits(binary.LittleEndian.Uint32(c.arena.buffer[c.offset+4 : c.offset+8]))
	return complex(r, i)
}

func (c Complex64) Set(v complex64) {
	binary.LittleEndian.PutUint32(c.arena.buffer[c.offset:c.offset+4], math.Float32bits(real(v)))
	binary.LittleEndian.PutUint32(c.arena.buffer[c.offset+4:c.offset+8], math.Float32bits(imag(v)))
}

// Complex128 wrapper
type Complex128 struct {
	arena  *Arena
	offset int
}

func (a *Arena) AllocComplex128() (Complex128, error) {
	offset, err := a.Alloc(16)
	if err != nil {
		return Complex128{}, err
	}
	return Complex128{arena: a, offset: offset}, nil
}

func (c Complex128) Get() complex128 {
	r := math.Float64frombits(binary.LittleEndian.Uint64(c.arena.buffer[c.offset : c.offset+8]))
	i := math.Float64frombits(binary.LittleEndian.Uint64(c.arena.buffer[c.offset+8 : c.offset+16]))
	return complex(r, i)
}

func (c Complex128) Set(v complex128) {
	binary.LittleEndian.PutUint64(c.arena.buffer[c.offset:c.offset+8], math.Float64bits(real(v)))
	binary.LittleEndian.PutUint64(c.arena.buffer[c.offset+8:c.offset+16], math.Float64bits(imag(v)))
}

// String wrapper
type String struct {
	arena  *Arena
	offset int
}

type Primitive interface {
	Int8 | Int16 | Int32 | Int64 |
		Uint8 | Uint16 | Uint32 | Uint64 |
		Float32 | Float64 |
		Complex64 | Complex128 |
		String
}

func (a *Arena) AllocString(v string) (String, error) {
	// 8 bytes for data offset, 8 bytes for length
	headerOffset, err := a.Alloc(16)
	if err != nil {
		return String{}, err
	}

	dataOffset, err := a.Alloc(len(v))
	if err != nil {
		return String{}, err
	}

	binary.LittleEndian.PutUint64(a.buffer[headerOffset:headerOffset+8], uint64(dataOffset))
	binary.LittleEndian.PutUint64(a.buffer[headerOffset+8:headerOffset+16], uint64(len(v)))
	copy(a.buffer[dataOffset:dataOffset+len(v)], v)

	return String{arena: a, offset: headerOffset}, nil
}

func (s String) Get() string {
	dataOffset := int(binary.LittleEndian.Uint64(s.arena.buffer[s.offset : s.offset+8]))
	length := int(binary.LittleEndian.Uint64(s.arena.buffer[s.offset+8 : s.offset+16]))
	return string(s.arena.buffer[dataOffset : dataOffset+length])
}

func (s String) Set(v string) error {
	dataOffset, err := s.arena.Alloc(len(v))
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint64(s.arena.buffer[s.offset:s.offset+8], uint64(dataOffset))
	binary.LittleEndian.PutUint64(s.arena.buffer[s.offset+8:s.offset+16], uint64(len(v)))
	copy(s.arena.buffer[dataOffset:dataOffset+len(v)], v)

	return nil
}

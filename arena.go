package arena

import "errors"

var ErrOutOfMemory = errors.New("out of memory")

// Arena represents a continuous block of byte memory
type Arena struct {
	buffer []byte
	offset int
	index  map[string]int
}

// New creates a new Arena with the specified byte size.
func New(size int) *Arena {
	return &Arena{
		buffer: make([]byte, size),
		offset: 0,
		index:  make(map[string]int),
	}
}

// Alloc reserves a block of 'size' bytes and returns the start offset.
func (a *Arena) Alloc(size int) (int, error) {
	if a.offset+size > len(a.buffer) {
		return 0, ErrOutOfMemory
	}
	start := a.offset
	a.offset += size
	return start, nil
}

// Reset clears the bump allocator offset, effectively freeing all memory.
// It also clears the index map. It does not shrink the buffer.
func (a *Arena) Reset() {
	a.offset = 0
	clear(a.index)
}

// SetName aliases a specific offset to a generic name for dictionary-based access.
func (a *Arena) SetName(name string, offset int) {
	if a.index == nil {
		a.index = make(map[string]int)
	}
	a.index[name] = offset
}

// GetName retrieves an offset mapped by string alias.
func (a *Arena) GetName(name string) (int, bool) {
	if a.index == nil {
		return 0, false
	}
	offset, ok := a.index[name]
	return offset, ok
}

// Buffer returns the underlying byte buffer.
func (a *Arena) Buffer() []byte {
	return a.buffer
}

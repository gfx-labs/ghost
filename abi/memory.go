package abi

// Memory is the underlying store for decoder data
type Memory interface {
	// returns the full data
	Data() []byte
	// increments the cursor by input, returns new cursor
	Pos(int) int
	// insert bytes to a location in memory, appending if needed
	Insert(loc int, data []byte)
	// replacing bytes at a location in memory, growing the slice if needed
	Put(loc int, data []byte)
}

// default sliceMemory implementation
type sliceMemory struct {
	encoded []byte // already encoded.
	cur     int    // current pointer (bytes)
}

func (m *sliceMemory) Data() []byte {
	return m.encoded[:m.cur]
}
func (m *sliceMemory) Pos(i int) int {
	m.cur = m.cur + i
	return m.cur
}

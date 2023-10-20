package abi

// inserts data at the location
func (m *sliceMemory) Insert(loc int, data []byte) {
	if loc == -1 {
		loc = m.cur
	}
	var s []byte
	if loc == 0 {
		s = data
	} else {
		s = append(m.encoded[:loc], data...)
	}
	if loc == len(m.encoded) {
		m.encoded = s
	} else {
		m.encoded = append(s, m.encoded[loc:]...)
	}
	m.Pos(len(data))
}

// replaces data at the location
func (m *sliceMemory) Put(loc int, data []byte) {
	if loc == -1 {
		loc = m.cur
	}
	if loc+len(data) > len(m.encoded) {
		m.grow(loc + len(data) - len(m.encoded))
	}
	copy(m.encoded[loc:loc+len(data)], data)
}

func (m *sliceMemory) grow(amt int) {
	m.encoded = append(m.encoded, make([]byte, amt)...)
	m.Pos(amt)
}

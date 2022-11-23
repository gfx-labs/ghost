package abi

func padleft(data []byte) []byte {
	alloc := [32]byte{}
	if len(data) == 0 {
		return alloc[:]
	}
	l := len(data) % 32
	if l == 0 {
		return data
	}
	padding := alloc[l:]

	return append(padding, data...)
}
func padright(data []byte) []byte {
	alloc := [32]byte{}
	if len(data) == 0 {
		return alloc[:]
	}
	l := len(data) % 32
	if l == 0 {
		return data
	}
	padding := alloc[l:]
	return append(data, padding...)
}

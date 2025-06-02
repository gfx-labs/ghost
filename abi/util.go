package abi

func padleft(data []byte) []byte {
	l := len(data) % 32
	if l == 0 && len(data) != 0 {
		return data
	}
	alloc := [32]byte{}
	padding := alloc[l:]

	return append(padding, data...)
}
func padright(data []byte) []byte {
	l := len(data) % 32
	if l == 0 && len(data) != 0 {
		return data
	}
	alloc := [32]byte{}
	padding := alloc[l:]
	return append(data, padding...)
}

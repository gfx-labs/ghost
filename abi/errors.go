package abi

import "errors"

var (
	// ErrUnexpectedEOF is returned when a read operation requires more
	// bytes than remain in the decoder.
	ErrUnexpectedEOF = errors.New("abi: unexpected EOF")
)

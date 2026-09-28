package abi

import "errors"

var (
	// ErrUnexpectedEOF is returned when a read operation requires more
	// bytes than remain in the decoder.
	ErrUnexpectedEOF = errors.New("abi: unexpected EOF")

	ErrOffsetOverflow  = errors.New("abi: dynamic offset overflow")
	ErrDynamicOverflow = errors.New("abi: dynamic overflow")
	ErrLenEOF          = errors.New("abi: len unexpected EOF")
	ErrLenOverflow     = errors.New("abi: len overflow")
)

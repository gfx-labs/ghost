// Package abi provides low-level encoding and decoding of Ethereum ABI data.
//
// The ABI (Application Binary Interface) is the standard way to interact with
// contracts in the Ethereum ecosystem. This package encodes Go values into
// ABI-compliant byte sequences and decodes them back.
//
// # Encoding
//
// Use [Builder] to construct ABI-encoded output. Builder uses a fluent API
// where each method returns the builder for chaining. Static types are written
// inline, while dynamic types (strings, bytes, slices) use Enter/Exit groups.
//
// Encoding a simple function call (transfer(address,uint256)):
//
//	b := new(abi.Builder)
//	b.Address(recipient).Uint256(amount)
//	calldata := b.Finish(abi.SIG("transfer", abi.ADDRESS, abi.UINT256).Fn())
//
// Encoding with dynamic types:
//
//	b := new(abi.Builder)
//	b.Uint(7).
//		EnterDynamicArray().          // begin uint256[]
//			Uint(1).Uint(2).Uint(3).
//		Exit().                       // end uint256[]
//		DString("hello")              // a dynamic string
//	encoded := b.Finish()
//
// Encoding nested tuples:
//
//	b := new(abi.Builder)
//	b.EnterTuple().
//		Address(addr).
//		EnterDynamicArray().
//			EnterTuple().Uint(1).Uint8(2).Exit().
//			EnterTuple().Uint(3).Uint8(4).Exit().
//		Exit().
//	Exit()
//	encoded := b.Finish()
//
// For fixed-size arrays of a known element type, use [Builder.EnterArray]:
//
//	b.EnterArray(abi.UINT256, 3).Uint(10).Uint(20).Uint(30).Exit()
//
// # Decoding
//
// Use [Decoder] to read ABI-encoded bytes sequentially. Each Read method
// consumes one 32-byte word (or follows dynamic offsets) and advances the cursor.
//
//	dec := abi.NewDecoder(data)
//	value, err := dec.Uint256()       // read a uint256
//	addr, err := dec.Address()        // read an address
//	str, err := dec.DString()         // read a dynamic string (follows offset)
//
// For dynamic arrays, use [Decoder.Dynamic] or [Decoder.DynamicLength] to follow
// the offset pointer, then read elements from the sub-decoder:
//
//	sub, length, err := dec.DynamicLength()
//	for i := 0; i < length; i++ {
//		val, err := sub.Uint()
//	}
//
// The decoder also supports non-advancing reads ([Decoder.Peek], [Decoder.PeekWord],
// [Decoder.PeekUint256]) and repositioning via [Decoder.Seek].
//
// # Type System
//
// [TypeName] represents Solidity type strings like "uint256", "address",
// "string", "(uint256,address)", "uint256[]", and "bytes32". Use the provided
// constants (UINT256, ADDRESS, BOOL, etc.) and constructors ([TUPLE], [SLICE],
// [ARRAY], [SIG]) to build type names programmatically.
//
// # Function Signatures
//
// [Signature] represents a canonical function signature string like
// "transfer(address,uint256)". Use [SIG] to build one from a method name
// and argument types. The [Signature.Hash] method computes the Keccak-256
// hash (cached), and [Signature.Fn] / [Signature.Selector] extract the
// 4-byte function selector used in calldata.
//
//	sig := abi.SIG("transfer", abi.ADDRESS, abi.UINT256)
//	selector := sig.Fn()    // first 4 bytes of keccak256("transfer(address,uint256)")
//	calldata := b.Finish(selector)
//
// # Memory
//
// Builder stores encoded data in a [Memory] implementation. The default
// is a simple byte-slice backend. Custom implementations can be injected
// via [WithBuilderMemory] for specialized use cases.
package abi

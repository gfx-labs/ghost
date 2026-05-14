// Package abir provides reflection-based ABI encoding and decoding,
// mapping Go structs to and from Ethereum ABI data using the [abi] package.
//
// While [abi.Builder] and [abi.Decoder] require manual field-by-field encoding,
// abir automates this via Go reflection. You provide the Go value and the ABI
// type descriptors, and abir handles the mapping.
//
// # Encoding
//
// Use [Encode] to serialize a Go value into an [abi.Builder]:
//
//	type Transfer struct {
//		To    string `abi:"address"`
//		Value uint
//	}
//	b := new(abi.Builder)
//	err := abir.Encode(b, Transfer{
//		To:    "0xdeadbeefcafebabedeadbeefcafebabedeadbeef",
//		Value: 1000,
//	}, abi.ADDRESS, abi.UINT)
//	calldata := b.Finish(abi.SIG("transfer", abi.ADDRESS, abi.UINT).Fn())
//
// With a single type argument, Encode writes the value directly. With multiple
// type arguments, the value must be a struct whose fields correspond to the
// types in order.
//
// # Decoding
//
// Use [Decode] to deserialize ABI data into Go values:
//
//	type Result struct {
//		Balance uint
//		Name    string
//	}
//	var r Result
//	err := abir.Decode(abi.NewDecoder(data), &r, abi.UINT, abi.STRING)
//
// [DecodeInto] infers ABI types from the struct's Go types and abi tags,
// so no explicit type list is needed:
//
//	type Balance struct {
//		Amount uint256.Int `abi:"uint256"`
//	}
//	var b Balance
//	err := abir.DecodeInto(decoder, &b)
//
// [DecodeBytes] is a convenience that wraps [Decode] with [abi.NewDecoder]:
//
//	var amount uint
//	err := abir.DecodeBytes(rawBytes, &amount, abi.UINT)
//
// # Struct Tags
//
// Fields can use the `abi` struct tag to override the inferred ABI type:
//
//	type Token struct {
//		Address string     `abi:"address"`    // encode string as address
//		Balance uint256.Int `abi:"uint256"`   // explicit type
//		Skip    int         `abi:"-"`         // excluded from encoding/decoding
//	}
//
// Without a tag, types are inferred via [CreateTypeName]:
//   - bool → "bool"
//   - uint, uint8..uint64 → "uint", "uint8".."uint64"
//   - int, int8..int64 → "int", "int8".."int64"
//   - string → "string"
//   - common.Address → "address"
//   - common.Hash → "bytes32"
//   - uint256.Int, *big.Int → "uint256"
//   - [N]byte (N ≤ 32) → "bytesN"
//   - []T → "T[]"
//   - [N]T → "T[N]"
//   - struct → "(field1Type,field2Type,...)"
//
// # Supported Target Types
//
// When decoding, the ABI type determines the source encoding, and the Go target
// type controls how the value is stored. The supported combinations include:
//
//   - Numeric types (uint/int) → Go int/uint variants, float32/64, big.Int, uint256.Int, [4]uint64, []uint64
//   - address → string (hex), []byte, [20]byte
//   - bool → bool, int/uint (0/1), string ("true"/"false", returns error)
//   - string → string, []uint8, []int32, [N]uint8
//   - bytes → string, []byte
//   - bytesN → string (hex), []byte, [N]byte, [N]int8
//
// # Error Handling
//
// Both Encode and Decode use recover() internally, so panics from the underlying
// abi package (e.g. from truncated data) are caught and returned as errors.
package abir

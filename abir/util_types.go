package abir

import (
	"math/big"
	"reflect"

	"github.com/holiman/uint256"
)

var typeUint256 = reflect.TypeOf(uint256.Int{})
var typeUint256Ptr = reflect.TypeOf(&uint256.Int{})
var typeBigInt = reflect.TypeOf(big.Int{})
var typeBigIntPtr = reflect.TypeOf(&big.Int{})

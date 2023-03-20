package abir

import (
	"math/big"
	"reflect"
)

var typeBigInt = reflect.TypeOf(big.Int{})
var typeBigIntPtr = reflect.TypeOf(&big.Int{})

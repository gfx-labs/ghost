package abir

import (
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var typeCommonHash = reflect.TypeOf(common.Hash{})
var typeCommonHashPtr = reflect.TypeOf(&common.Hash{})
var typeCommonAddress = reflect.TypeOf(common.Address{})
var typeCommonAddressPtr = reflect.TypeOf(&common.Address{})

var typeUint8 = reflect.TypeOf(uint8(0))
var typeUint256 = reflect.TypeOf(uint256.Int{})
var typeUint256Ptr = reflect.TypeOf(&uint256.Int{})
var typeBigInt = reflect.TypeOf(big.Int{})
var typeBigIntPtr = reflect.TypeOf(&big.Int{})

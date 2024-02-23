package ethtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)

type Transaction struct {
	GasPrice         uint256.Int     `json:"gasPrice"`
	ChainID          hexutil.Uint64  `json:"chainId"`
	BlockHash        common.Hash     `json:"blockHash"`
	Type             hexutil.Uint64  `json:"type"`
	Gas              uint256.Int     `json:"gas"`
	S                uint256.Int     `json:"s"`
	From             common.Address  `json:"from"`
	Hash             common.Hash     `json:"hash"`
	TransactionIndex hexutil.Uint64  `json:"transactionIndex"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	Input            hexutil.Bytes   `json:"input"`
	BlockNumber      hexutil.Uint64  `json:"blockNumber"`
	To               *common.Address `json:"to"`
	V                uint256.Int     `json:"v"`
	R                uint256.Int     `json:"r"`
	Value            uint256.Int     `json:"value"`

	AccessList           []AccessListItem `json:"accessList,omitempty"`
	MaxFeePerGas         uint256.Int      `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas uint256.Int      `json:"maxPriorityFeePerGas,omitempty"`
}

type AccessListItem struct {
	Address     common.Address `json:"address"`
	StorageKeys []common.Hash  `json:"storageKeys"`
}

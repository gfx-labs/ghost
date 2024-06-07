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

type Log struct {
	Address          common.Address `json:"address"`
	Topics           []common.Hash  `json:"topics"`
	Data             hexutil.Bytes  `json:"data"`
	BlockNumber      hexutil.Uint64 `json:"blockNumber"`
	TransactionHash  common.Hash    `json:"transactionHash"`
	TransactionIndex hexutil.Uint64 `json:"transactionIndex"`
	BlockHash        common.Hash    `json:"blockHash"`
	LogIndex         hexutil.Uint64 `json:"logIndex"`
	Removed          bool           `json:"removed"`
}

type TransactionReceipt struct {
	BlockHash         common.Hash    `json:"blockHash"`
	BlockNumber       hexutil.Uint64 `json:"blockNumber"`
	ContractAddress   common.Address `json:"contractAddress"`
	CumulativeGasUsed uint256.Int    `json:"cumulativeGasUsed"`
	EffectiveGasPrice uint256.Int    `json:"effectiveGasPrice"`
	From              common.Address `json:"from"`
	GasUsed           uint256.Int    `json:"gasUsed"`
	Logs              []Log          `json:"logs"`
	LogsBloom         hexutil.Bytes  `json:"logsBloom"`
	Status            hexutil.Uint64 `json:"status"`
	To                common.Address `json:"to"`
	TransactionHash   common.Hash    `json:"transactionHash"`
	TransactionIndex  hexutil.Uint64 `json:"transactionIndex"`
	Type              hexutil.Uint64 `json:"type"`
}

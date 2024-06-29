package ethtypes

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)

type Transaction struct {
	GasPrice         uint256.Int     `json:"gasPrice"`
	ChainID          hexutil.Uint64  `json:"chainId"`
	BlockHash        *common.Hash    `json:"blockHash"`
	Type             hexutil.Uint64  `json:"type"`
	Gas              uint256.Int     `json:"gas"`
	S                uint256.Int     `json:"s"`
	From             common.Address  `json:"from"`
	Hash             common.Hash     `json:"hash"`
	TransactionIndex *hexutil.Uint64 `json:"transactionIndex"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	Input            hexutil.Bytes   `json:"input"`
	BlockNumber      *hexutil.Uint64 `json:"blockNumber"`
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
	Address          common.Address  `json:"address"`
	Topics           []hexutil.Bytes `json:"topics"`
	Data             hexutil.Bytes   `json:"data"`
	BlockNumber      hexutil.Uint64  `json:"blockNumber"`
	TransactionHash  *common.Hash    `json:"transactionHash"`
	TransactionIndex *hexutil.Uint64 `json:"transactionIndex"`
	BlockHash        common.Hash     `json:"blockHash"`
	LogIndex         *hexutil.Uint64 `json:"logIndex"`
	Removed          bool            `json:"removed"`
}

type TransactionReceipt struct {
	BlockHash         common.Hash     `json:"blockHash"`
	BlockNumber       hexutil.Uint64  `json:"blockNumber"`
	ContractAddress   *common.Address `json:"contractAddress"`
	CumulativeGasUsed uint256.Int     `json:"cumulativeGasUsed"`
	EffectiveGasPrice uint256.Int     `json:"effectiveGasPrice"`
	From              common.Address  `json:"from"`
	GasUsed           uint256.Int     `json:"gasUsed"`
	Logs              []Log           `json:"logs"`
	LogsBloom         hexutil.Bytes   `json:"logsBloom"`
	Status            hexutil.Uint64  `json:"status"`
	To                *common.Address `json:"to"`
	TransactionHash   common.Hash     `json:"transactionHash"`
	TransactionIndex  hexutil.Uint64  `json:"transactionIndex"`
	Type              hexutil.Uint64  `json:"type"`
}

type Block struct {
	BaseFeePerGas    uint256.Int       `json:"baseFeePerGas,omitempty"`
	Difficulty       uint256.Int       `json:"difficulty"`
	ExtraData        hexutil.Bytes     `json:"extraData"`
	GasLimit         uint256.Int       `json:"gasLimit"`
	GasUsed          uint256.Int       `json:"gasUsed"`
	Hash             *common.Hash      `json:"hash"`
	LogsBloom        *hexutil.Bytes    `json:"logsBloom"`
	Miner            common.Address    `json:"miner"`
	MixHash          common.Hash       `json:"mixHash"`
	Nonce            *hexutil.Bytes    `json:"nonce"` // always 8 bytes
	Number           *hexutil.Uint64   `json:"blockNumber"`
	ParentHash       common.Hash       `json:"parentHash"`
	ReceiptsRoot     common.Hash       `json:"receiptsRoot"`
	Sha3Uncles       common.Hash       `json:"sha3Uncles"`
	Size             uint256.Int       `json:"size"`
	StateRoot        common.Hash       `json:"stateRoot"`
	Timestamp        hexutil.Uint64    `json:"timestamp"`
	TotalDifficulty  uint256.Int       `json:"totalDifficulty"`
	TransactionsRoot common.Hash       `json:"transactionsRoot"`
	Uncles           []common.Hash     `json:"uncles"`
	Withdrawals      []BlockWithdrawal `json:"withdrawals,omitempty"`
	WithdrawalsRoot  *common.Hash      `json:"withdrawalsRoot,omitempty"`
}

type BlockWithdrawal struct {
	Amount         uint256.Int    `json:"amount"`
	Address        common.Address `json:"address"`
	Index          hexutil.Uint64 `json:"index"`
	ValidatorIndex hexutil.Uint64 `json:"validatorIndex"`
}

type BlockTxHashes struct {
	Block
	Transactions []common.Hash `json:"transactions"`
}

type BlockTxObjs struct {
	Block
	Transactions []Transaction `json:"transactions"`
}

type Trace struct {
	Type         string         `json:"type"`
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Value        uint256.Int    `json:"value"`
	Gas          uint256.Int    `json:"gas"`
	GasUsed      uint256.Int    `json:"gasUsed"`
	Input        hexutil.Bytes  `json:"input"`
	Output       *hexutil.Bytes `json:"output,omitempty"`
	Calls        []Trace        `json:"calls,omitempty"`
	Error        *string        `json:"error,omitempty"`
	RevertReason *string        `json:"revertReason,omitempty"`
}

//---------------------------- ABI TYPES

type AbiComponent struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Components []AbiComponent `json:"components"`
}

type AbiParam struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Components []AbiComponent `json:"components"`
	Indexed    bool           `json:"indexed"` // only for event args
}

type AbiEventFn struct { //event or function or error
	Type            string     `json:"type"`
	Name            string     `json:"name"`
	Inputs          []AbiParam `json:"inputs"`
	Outputs         []AbiParam `json:"outputs"`         // doesnt exist for events
	Anonymous       bool       `json:"anonymous"`       // only for events
	StateMutability string     `json:"stateMutability"` // only for functions
}

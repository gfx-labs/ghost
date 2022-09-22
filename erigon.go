package ghost

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type ErigonClient interface {
	ErigonFilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]ErigonLog, error)
	ErigonGetReceiptsByHash(ctx context.Context, txHash common.Hash) ([]ErigonReceipt, error)

	Client
}

type BlockReceipt struct {
	BlockNumber uint64      `json:"blockNumber,omitempty"`
	BlockHash   common.Hash `json:"blockHash,omitempty"`

	TransactionHash common.Hash     `json:"omitempty"`
	ContractAddress *common.Address `json:"contractAddress,omitempty"`

	ChainID string `json:"chainId,omitempty"`

	From             common.Address   `json:"from,omitempty"`
	To               common.Address   `json:"to,omitempty"`
	Input            hexutil.Bytes    `json:"input,omitempty"`
	TransactionIndex uint64           `json:"transactionIndex,omitempty"`
	Value            hexutil.Big      `json:"value,omitempty"`
	AccessList       types.AccessList `json:"accessList,omitempty"`

	Status  string      `json:"status,omitempty"`
	GasUsed hexutil.Big `json:"gasUsed,omitempty"`

	Type uint64 `json:"type,omitempty"`

	Gas                  uint64       `json:"gas,omitempty"`
	GasPrice             *hexutil.Big `json:"gasPrice,omitempty"`
	MaxFeePerGas         *hexutil.Big `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *hexutil.Big `json:"maxPriorityFeePerGas,omitempty"`

	CumulativeGasUsed hexutil.Big `json:"cumulativeGasUsed,omitempty"`
	EffectiveGasPrice hexutil.Big `json:"effectiveGasPrice,omitempty"`

	Nonce uint64       `json:"nonce,omitempty"`
	R     *hexutil.Big `json:"r,omitempty"`
	S     *hexutil.Big `json:"s,omitempty"`
	V     *hexutil.Big `json:"v,omitempty"`

	LogsBloom hexutil.Bytes `json:"logsBloom,omitempty"`
	Logs      []*ErigonLog  `json:"logs,omitempty"`
}

type ErigonLog struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required" codec:"1"`
	// list of topics provided by the contract.
	Topics []common.Hash `json:"topics" gencodec:"required" codec:"2"`
	// supplied by the contract, usually ABI-encoded
	Data []byte `json:"data" gencodec:"required" codec:"3"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockNumber uint64 `json:"blockNumber" codec:"-"`

	Timestamp uint64 `json:"timestamp" codec:"-"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required" codec:"-"`
	// index of the transaction in the block
	TxIndex uint `json:"transactionIndex" codec:"-"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash" codec:"-"`
	// index of the log in the block
	Index uint `json:"logIndex" codec:"-"`

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed" codec:"-"`
}

// MarshalJSON marshals as JSON.
func (l ErigonLog) MarshalJSON() ([]byte, error) {
	type Log struct {
		Address     common.Address `json:"address" gencodec:"required"`
		Topics      []common.Hash  `json:"topics" gencodec:"required"`
		Data        hexutil.Bytes  `json:"data" gencodec:"required"`
		BlockNumber hexutil.Uint64 `json:"blockNumber"`
		Timestamp   hexutil.Uint64 `json:"timestamp"`
		TxHash      common.Hash    `json:"transactionHash" gencodec:"required"`
		TxIndex     hexutil.Uint   `json:"transactionIndex"`
		BlockHash   common.Hash    `json:"blockHash"`
		Index       hexutil.Uint   `json:"logIndex"`
		Removed     bool           `json:"removed"`
	}
	var enc Log
	enc.Address = l.Address
	enc.Topics = l.Topics
	enc.Data = l.Data
	enc.BlockNumber = hexutil.Uint64(l.BlockNumber)
	enc.Timestamp = hexutil.Uint64(l.Timestamp)
	enc.TxHash = l.TxHash
	enc.TxIndex = hexutil.Uint(l.TxIndex)
	enc.BlockHash = l.BlockHash
	enc.Index = hexutil.Uint(l.Index)
	enc.Removed = l.Removed
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (l *ErigonLog) UnmarshalJSON(input []byte) error {
	type Log struct {
		Address     *common.Address `json:"address" gencodec:"required"`
		Topics      []common.Hash   `json:"topics" gencodec:"required"`
		Data        *hexutil.Bytes  `json:"data" gencodec:"required"`
		BlockNumber *hexutil.Uint64 `json:"blockNumber"`
		Timestamp   *hexutil.Uint64 `json:"timestamp"`
		TxHash      *common.Hash    `json:"transactionHash" gencodec:"required"`
		TxIndex     *hexutil.Uint   `json:"transactionIndex"`
		BlockHash   *common.Hash    `json:"blockHash"`
		Index       *hexutil.Uint   `json:"logIndex"`
		Removed     *bool           `json:"removed"`
	}
	var dec Log
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Address == nil {
		return errors.New("missing required field 'address' for Log")
	}
	l.Address = *dec.Address
	if dec.Topics == nil {
		return errors.New("missing required field 'topics' for Log")
	}
	l.Topics = dec.Topics
	if dec.Data == nil {
		return errors.New("missing required field 'data' for Log")
	}
	l.Data = *dec.Data
	if dec.BlockNumber != nil {
		l.BlockNumber = uint64(*dec.BlockNumber)
	}
	if dec.Timestamp != nil {
		l.Timestamp = uint64(*dec.Timestamp)
	}
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Log")
	}
	l.TxHash = *dec.TxHash
	if dec.TxIndex != nil {
		l.TxIndex = uint(*dec.TxIndex)
	}
	if dec.BlockHash != nil {
		l.BlockHash = *dec.BlockHash
	}
	if dec.Index != nil {
		l.Index = uint(*dec.Index)
	}
	if dec.Removed != nil {
		l.Removed = *dec.Removed
	}
	return nil
}

type ErigonReceipt struct {
	// Consensus fields: These fields are defined by the Yellow Paper
	Type              uint8          `json:"type,omitempty"`
	PostState         []byte         `json:"root"`
	Status            uint64         `json:"status"`
	CumulativeGasUsed uint64         `json:"cumulativeGasUsed" gencodec:"required"`
	Bloom             *hexutil.Bytes `json:"logsBloom"         gencodec:"required"`
	Logs              []*types.Log   `json:"logs"              gencodec:"required"`

	// Implementation fields: These fields are added by geth when processing a transaction.
	// They are stored in the chain database.
	TxHash          common.Hash    `json:"transactionHash" gencodec:"required"`
	ContractAddress common.Address `json:"contractAddress"`
	GasUsed         uint64         `json:"gasUsed" gencodec:"required"`

	// Inclusion information: These fields provide information about the inclusion of the
	// transaction corresponding to this receipt.
	BlockHash        common.Hash `json:"blockHash,omitempty"`
	BlockNumber      *big.Int    `json:"blockNumber,omitempty"`
	TransactionIndex uint        `json:"transactionIndex"`
}

// MarshalJSON marshals as JSON.
func (r ErigonReceipt) MarshalJSON() ([]byte, error) {
	type Receipt struct {
		Type              hexutil.Uint64 `json:"type,omitempty"`
		PostState         hexutil.Bytes  `json:"root"`
		Status            hexutil.Uint64 `json:"status"`
		CumulativeGasUsed hexutil.Uint64 `json:"cumulativeGasUsed" gencodec:"required"`
		Bloom             *hexutil.Bytes `json:"logsBloom"         gencodec:"required"`
		Logs              []*types.Log   `json:"logs"              gencodec:"required"`
		TxHash            common.Hash    `json:"transactionHash" gencodec:"required"`
		ContractAddress   common.Address `json:"contractAddress"`
		GasUsed           hexutil.Uint64 `json:"gasUsed" gencodec:"required"`
		BlockHash         common.Hash    `json:"blockHash,omitempty"`
		BlockNumber       *hexutil.Big   `json:"blockNumber,omitempty"`
		TransactionIndex  hexutil.Uint   `json:"transactionIndex"`
	}
	var enc Receipt
	enc.Type = hexutil.Uint64(r.Type)
	enc.PostState = r.PostState
	enc.Status = hexutil.Uint64(r.Status)
	enc.CumulativeGasUsed = hexutil.Uint64(r.CumulativeGasUsed)
	enc.Bloom = r.Bloom
	enc.Logs = r.Logs
	enc.TxHash = r.TxHash
	enc.ContractAddress = r.ContractAddress
	enc.GasUsed = hexutil.Uint64(r.GasUsed)
	enc.BlockHash = r.BlockHash
	enc.BlockNumber = (*hexutil.Big)(r.BlockNumber)
	enc.TransactionIndex = hexutil.Uint(r.TransactionIndex)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (r *ErigonReceipt) UnmarshalJSON(input []byte) error {
	type Receipt struct {
		Type              *hexutil.Uint64 `json:"type,omitempty"`
		PostState         *hexutil.Bytes  `json:"root"`
		Status            *hexutil.Uint64 `json:"status"`
		CumulativeGasUsed *hexutil.Uint64 `json:"cumulativeGasUsed" gencodec:"required"`
		Bloom             hexutil.Bytes   `json:"logsBloom"         `
		Logs              []*types.Log    `json:"logs"              `
		TxHash            *common.Hash    `json:"transactionHash" gencodec:"required"`
		ContractAddress   *common.Address `json:"contractAddress"`
		GasUsed           *hexutil.Uint64 `json:"gasUsed" gencodec:"required"`
		BlockHash         *common.Hash    `json:"blockHash,omitempty"`
		BlockNumber       *hexutil.Big    `json:"blockNumber,omitempty"`
		TransactionIndex  *hexutil.Uint   `json:"transactionIndex"`
	}
	var dec Receipt
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Type != nil {
		r.Type = uint8(*dec.Type)
	}
	if dec.PostState != nil {
		r.PostState = *dec.PostState
	}
	if dec.Status != nil {
		r.Status = uint64(*dec.Status)
	}
	if dec.CumulativeGasUsed == nil {
		return errors.New("missing required field 'cumulativeGasUsed' for Receipt")
	}
	r.CumulativeGasUsed = uint64(*dec.CumulativeGasUsed)
	if dec.Bloom != nil {
		r.Bloom = &dec.Bloom
	}
	r.Logs = dec.Logs
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Receipt")
	}
	r.TxHash = *dec.TxHash
	if dec.ContractAddress != nil {
		r.ContractAddress = *dec.ContractAddress
	}
	if dec.GasUsed == nil {
		return errors.New("missing required field 'gasUsed' for Receipt")
	}
	r.GasUsed = uint64(*dec.GasUsed)
	if dec.BlockHash != nil {
		r.BlockHash = *dec.BlockHash
	}
	if dec.BlockNumber != nil {
		r.BlockNumber = (*big.Int)(dec.BlockNumber)
	}
	if dec.TransactionIndex != nil {
		r.TransactionIndex = uint(*dec.TransactionIndex)
	}
	return nil
}

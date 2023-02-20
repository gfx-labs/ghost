package evm

import (
	"math/big"

	"gfx.cafe/open/ghost/ethtype"
	"gfx.cafe/open/ghost/evm/bloom"
)

type Address = ethtype.Address

type Account interface {
}

type Contract interface {
	Address() Address
	StorageAt(Word) Word
	WriteStorage(Word, Word) error
}

// Message represents a message sent to a contract.
type Message interface {
	From() Address
	To() *Address

	GasPrice() Word
	GasFeeCap() Word
	GasTipCap() Word
	Gas() uint64
	Value() Word

	Nonce() uint64
	IsFake() bool
	Data() []byte
	AccessList() AccessList
}

// Engine is an interface used to execute a stack
// is also is what holds all the memory and call data, and is ultimately responsible for
// obtaining the post execution state
type Engine interface {
	CreateContract() error
	// the calldata. returns nil when empty
	CallData(idx uint64) Word
	CallDataSize() Word
	CallDataCopy(a, b, c Word) error

	CodeSize() Word
	CodeCopy(a, b, c Word) error
	ReturnDataSize() Word
	ReturnDataCopy(a, b, c Word) error
	ExtCodeSize(Address) Word
	ExtCodeCopy(a, b, c, d Word) error

	MemorySize() Word
	MemoryAt(Word) Word
	WriteMemory(Word, Word) error
	WriteMemoryByte(Word, byte) error

	Jump(Word) error
	Jump1(Word, Word) error
	Counter() Word

	WriteLog(offset Word, length Word, topics []Word) error
	Create(a, b, c Word) (Address, error)
	Create2(a, b, c, d Word) (Address, error)
	DelegateCall(a, b, c, d, e, f Word) (Word, error)
	StaticCall(a, b, c, d, e, f Word) (Word, error)
	Call(a, b, c, d, e, f, g Word) (Word, error)
	CallCode(a, b, c, d, e, f, g Word) (Word, error)
	ExtCodeHash(Address) Word

	Return() error

	Caller() Address
	Contract() Contract
	// get balance of account
	Balance(target Address) Word

	// the following fields should be passed to the execution context
	Msg() Message
	// input 0 should return the current block
	// input 1 should return the previous block
	Block(Word) Block
}

type Block interface {
	Coinbase() Address
	Hash() Word
	Difficulty() *big.Int
	Number() *big.Int
	GasLimit() uint64
	GasUsed() uint64
	Time() uint64
	ChainID() Word
}

type Receipt struct {
	// Consensus fields: These fields are defined by the Yellow Paper
	Type              uint8       `json:"type,omitempty"`
	PostState         []byte      `json:"root"`
	Status            uint64      `json:"status"`
	CumulativeGasUsed uint64      `json:"cumulativeGasUsed" gencodec:"required"`
	Bloom             bloom.Bloom `json:"logsBloom"         gencodec:"required"`
	Logs              []*Log      `json:"logs"              gencodec:"required"`

	TxHash          Word    `json:"transactionHash" gencodec:"required"`
	ContractAddress Address `json:"contractAddress"`
	GasUsed         uint64  `json:"gasUsed" gencodec:"required"`

	BlockHash        Word     `json:"blockHash,omitempty"`
	BlockNumber      *big.Int `json:"blockNumber,omitempty"`
	TransactionIndex uint     `json:"transactionIndex"`
}

type Log struct {
	Address Address `json:"address"`
	Topics  []Word  `json:"topics"`
	Data    []byte  `json:"data"`

	BlockNumber uint64 `json:"blockNumber"`
	TxHash      Word   `json:"transactionHash"`
	TxIndex     uint   `json:"transactionIndex"`
	BlockHash   Word   `json:"blockHash"`
	Index       uint   `json:"logIndex"`

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed"`
}

type AccessList []AccessTuple

// AccessTuple is the element type of an access list.
type AccessTuple struct {
	Address     Address `json:"address"        gencodec:"required"`
	StorageKeys []Word  `json:"storageKeys"    gencodec:"required"`
}

type Receipts []*Receipt

// CreateBloom creates a bloom filter out of the give Receipts (+Logs)
func CreateBloom(receipts Receipts) bloom.Bloom {
	buf := make([]byte, 6)
	var bin bloom.Bloom
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			bin.AddInternal(log.Address[:], buf)
			for _, b := range log.Topics {
				bin.AddInternal(b.Bytes(), buf)
			}
		}
	}
	return bin
}

// LogsBloom returns the bloom bytes for the given logs
func LogsBloom(logs []*Log) []byte {
	buf := make([]byte, 6)
	var bin bloom.Bloom
	for _, log := range logs {
		bin.AddInternal(log.Address[:], buf)
		for _, b := range log.Topics {
			bin.AddInternal(b.Bytes(), buf)
		}
	}
	return bin[:]
}

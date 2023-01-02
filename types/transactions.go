package types

import "math/big"

// A data model that represents a transaction.
type Tx struct {
	// keccab32 of the serialized byte. For utxo, it's the keccab32 hash of tx hash and utxo index.
	Hash        string
	OutputIndex int // used only for utxos
	Serialized  []byte
	From        string
	To          string
	Success     bool
}

// List of all transactions in a block of a specific chain.
type Txs struct {
	Chain     string
	Block     int64
	BlockHash string
	Arr       []*Tx

	// ETH only
	BaseFee     *big.Int
	PriorityFee *big.Int
}

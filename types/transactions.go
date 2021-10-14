package types

// A data model that represents a transaction.
type Tx struct {
	Hash       string
	Serialized []byte
	To         string
	From       string
}

// List of all transactions in a block of a specific chain.
type Txs struct {
	Chain string
	Block int64

	Arr []*Tx
}

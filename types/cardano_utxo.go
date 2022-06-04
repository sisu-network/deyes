package types

import "github.com/echovl/cardano-go"

// Represents a utxo transaction from address A -> B in Cardano.
type CardanoUtxo struct {
	TxHash  cardano.Hash32
	Spender cardano.Address
	Amount  *cardano.Value
	Index   uint64
}

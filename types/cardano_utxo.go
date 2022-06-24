package types

import (
	"fmt"
	"strconv"

	"github.com/echovl/cardano-go"
)

// Represents a utxo transaction from address A -> B in Cardano.
type CardanoUtxo struct {
	TxHash  cardano.Hash32
	Spender cardano.Address
	Amount  *cardano.Value
	Index   uint64
}

type CardanoTxInItem struct {
	TxHash cardano.Hash32
	// To Cardano gateway address
	To        cardano.Address
	UtxoIndex int

	// Info about the swap
	Amount   uint64
	Asset    string
	Metadata CardanoTxMetadata
}

type CardanoTxMetadata struct {
	// Get from transaction metadata
	Chain     string `json:"chain,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	// NativeAda indicates if user wants to transfer ADA token cross chain since every multi-asset
	// transaction requires some ADA in it.
	NativeAda int `json:"native_ada,omitempty"`
}

func (c *CardanoUtxo) Hash() string {
	return fmt.Sprintf("%s__%s", c.TxHash, strconv.Itoa(int(c.Index)))
}

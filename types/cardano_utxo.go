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
	// Recipient Cardano gateway address
	Recipient      cardano.Address
	TxAdditionInfo *TxAdditionInfo
}

type TxAdditionInfo struct {
	// Get from transaction metadata
	Chain        string `json:"chain,omitempty"`
	Recipient    string `json:"recipient,omitempty"`
	TokenAddress string `json:"token_address,omitempty"`

	// Get from transaction output
	Amount *cardano.Value `json:"amount,omitempty"`
}

func (tx *TxAdditionInfo) WithAmount(value *cardano.Value) *TxAdditionInfo {
	tx.Amount = value
	return tx
}

func (c *CardanoUtxo) Hash() string {
	return fmt.Sprintf("%s__%s", c.TxHash, strconv.Itoa(int(c.Index)))
}

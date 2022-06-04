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

func (c *CardanoUtxo) Hash() string {
	return fmt.Sprintf("%s__%s", c.TxHash, strconv.Itoa(int(c.Index)))
}

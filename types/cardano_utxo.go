package types

import (
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
)

///////////////////////////////////////////////////////////////

type CardanoTransactionUtxo struct {
	Hash     string                   `json:"hash"`
	Index    int                      `json:"index"`
	Address  string                   `json:"Address"`
	Amount   []providertypes.TxAmount `json:"amount"`
	Metadata *CardanoTxMetadata       `json:"metadata"`
}

type CardanoTxMetadata struct {
	// Get from transaction metadata
	Chain     string `json:"chain,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	// NativeAda indicates if user wants to transfer ADA token cross chain since every multi-asset
	// transaction requires some ADA in it.
	NativeAda int `json:"native_ada,omitempty"`
}

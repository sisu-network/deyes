package types

type TxAmount struct {
	// The quantity of the unit
	Quantity string `json:"quantity"`

	// The unit of the value
	Unit string `json:"unit"`
}

type CardanoTransactionUtxo struct {
	Hash     string             `json:"hash"`
	Index    int                `json:"index"`
	Address  string             `json:"Address"`
	Amount   []TxAmount         `json:"amount"`
	Metadata *CardanoTxMetadata `json:"metadata"`
}

type CardanoTxMetadata struct {
	// Get from transaction metadata
	Chain     string `json:"chain,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	// NativeAda indicates if user wants to transfer ADA token cross chain since every multi-asset
	// transaction requires some ADA in it.
	NativeAda int `json:"native_ada,omitempty"`
}

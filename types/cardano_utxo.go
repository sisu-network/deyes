package types

type TxAmount struct {
	// The quantity of the unit
	Quantity string `json:"quantity"`

	// The unit of the value
	Unit string `json:"unit"`
}

type TransactionUTXOs struct {
	// Transaction hash
	Hash   string `json:"hash"`
	Inputs []struct {
		// Input address
		Address string     `json:"address"`
		Amount  []TxAmount `json:"amount"`

		// UTXO index in the transaction
		OutputIndex float32 `json:"output_index"`

		// Hash of the UTXO transaction
		TxHash string `json:"tx_hash"`
	} `json:"inputs"`
	Outputs []struct {
		// Output address
		Address string     `json:"address"`
		Amount  []TxAmount `json:"amount"`
	} `json:"outputs"`
}

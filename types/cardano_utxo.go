package types

type CardanoBlock struct {
	// Block creation time in UNIX time
	Time int `json:"time"`

	// Block number
	Height int `json:"height"`

	// Hash of the block
	Hash string `json:"hash"`

	// Epoch number
	Epoch int `json:"epoch"`

	// Slot number
	Slot int `json:"slot"`
}

type AddressTransactions struct {
	// Hash of the transaction
	TxHash string `json:"tx_hash"`
}

type TransactionMetadata struct {
	// Content of the metadata
	JsonMetadata interface{} `json:"json_metadata"`

	// Metadata label
	Label string `json:"label"`
}

type TransactionUTXOsOutput struct {
	// Output address
	Address string     `json:"address"`
	Amount  []TxAmount `json:"amount"`
}

type TransactionUTXOs struct {
	// Transaction hash
	Hash    string                   `json:"hash"`
	Outputs []TransactionUTXOsOutput `json:"outputs"`
}

type TxAmount struct {
	// The quantity of the unit
	Quantity string `json:"quantity"`

	// The unit of the value
	Unit string `json:"unit"`
}

type EpochParameters struct {
	// Epoch number
	Epoch int `json:"epoch"`
	// The amount of a key registration deposit in Lovelaces
	KeyDeposit string `json:"key_deposit"`

	// Maximum block header size
	MaxBlockHeaderSize int `json:"max_block_header_size"`

	// Maximum block body size in Bytes
	MaxBlockSize int `json:"max_block_size"`

	// Maximum transaction size
	MaxTxSize int `json:"max_tx_size"`

	// The linear factor for the minimum fee calculation for given epoch
	MinFeeA int `json:"min_fee_a"`

	// The constant factor for the minimum fee calculation
	MinFeeB int `json:"min_fee_b"`

	// Minimum UTXO value
	MinUtxo string `json:"min_utxo"`

	// Desired number of pools
	NOpt int `json:"n_opt"`

	// The amount of a pool registration deposit in Lovelaces
	PoolDeposit string `json:"pool_deposit"`
}

type AddressAmount struct {
	// The quantity of the unit
	Quantity string `json:"quantity"`

	// The unit of the value
	Unit string `json:"unit"`
}

type AddressUTXO struct {
	// Transaction hash of the UTXO
	TxHash string `json:"tx_hash"`

	// UTXO index in the transaction
	OutputIndex int             `json:"output_index"`
	Amount      []AddressAmount `json:"amount"`

	// Block hash of the UTXO
	Block string `json:"block"`
}

// APIQueryParams contains query parameters. Marshalled to
// "count", "page", "order", "from", "to".
type APIQueryParams struct {
	Count int
	Page  int
	Order string
	From  string
	To    string
}

///////////////////////////////////////////////////////////////

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

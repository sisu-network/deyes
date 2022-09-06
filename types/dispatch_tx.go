package types

type DispatchedTxRequest struct {
	Chain  string
	Tx     []byte
	TxHash string

	// For ETH chains
	PubKey []byte
}

type DispatchedTxResult struct {
	Success bool
	Err     DispatchError // We use int since json RPC cannot marshal error
	Chain   string
	TxHash  string
}

func NewDispatchTxError(err DispatchError) *DispatchedTxResult {
	return &DispatchedTxResult{
		Success: false,
		Err:     err,
	}
}

package types

type DispatchedTxRequest struct {
	Chain  string
	Tx     []byte
	TxHash string

	// For ETH chains
	PubKey                  []byte
	IsEthContractDeployment bool
}

type DispatchedTxResult struct {
	Success bool
	Err     error
	Chain   string
	TxHash  string

	// For ETH only. This is optional deployed contract addresses.
	DeployedAddr            string
	IsEthContractDeployment bool
}

func NewDispatchTxError(err error) *DispatchedTxResult {
	return &DispatchedTxResult{
		Success: false,
		Err:     err,
	}
}

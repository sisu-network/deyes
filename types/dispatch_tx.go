package types

type DispatchedTxRequest struct {
	Chain string
	Tx    []byte

	// For ETH chains
	PubKey                  []byte
	IsEthContractDeployment bool
}

type DispatchedTxResult struct {
	Success bool
	Err     error

	// For ETH only. This is optional deployed contract addresses.
	DeployedAddr string
}

func NewDispatchTxError(err error) *DispatchedTxResult {
	return &DispatchedTxResult{
		Success: false,
		Err:     err,
	}
}

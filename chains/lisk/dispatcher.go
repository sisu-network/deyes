package lisk

import (
	"encoding/hex"

	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type LiskDispatcher struct {
	chain  string
	client Client
}

func NewDispatcher(chain string, client Client) chains.Dispatcher {
	return &LiskDispatcher{
		client: client,
		chain:  chain,
	}
}

func (d *LiskDispatcher) Start() {
}

func (d *LiskDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	txHash, err := d.client.CreateTransaction(hex.EncodeToString(request.Tx))
	if err != nil {
		log.Errorf("Failed to create transaction, err = %v", err)
		return types.NewDispatchTxError(request, types.ErrSubmitTx)
	}

	log.Infof("Returned lisk tx hash from server = %s", txHash)

	if len(txHash) == 0 {
		log.Errorf("failed to dispatch lisk transaction, server rejects the tx")
		return types.NewDispatchTxError(request, types.ErrGeneric)
	}

	return &types.DispatchedTxResult{
		Success: true,
		Chain:   request.Chain,
		TxHash:  txHash,
	}
}

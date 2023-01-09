package lisk

import (
	"encoding/hex"

	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type LiskDispatcher struct {
	chain  string
	client LiskClient
}

func NewDispatcher(chain string, client LiskClient) chains.Dispatcher {
	return &LiskDispatcher{
		client: client,
		chain:  chain,
	}
}

func (d *LiskDispatcher) Start() {

}

func (d *LiskDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	tx, err := d.client.CreateTransaction(hex.EncodeToString(request.Tx))
	if err != nil {
		log.Errorf("Failed to create transaction, err = %v", err)
		return &types.DispatchedTxResult{Success: false}
	}
	return &types.DispatchedTxResult{
		Success: true,
		Chain:   request.Chain,
		TxHash:  tx,
	}
}

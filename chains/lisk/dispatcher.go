package lisk

import (
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
)

type LiskDispatcher struct {
	client LiskClient
}

func NewDispatcher(client LiskClient) chains.Dispatcher {
	return &LiskDispatcher{
		client: client,
	}
}

func (d *LiskDispatcher) Start() {
}

func (d *LiskDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	return &types.DispatchedTxResult{
		Success: true,
		Chain:   request.Chain,
		TxHash:  "",
	}
}

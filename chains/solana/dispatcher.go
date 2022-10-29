package solana

import "github.com/sisu-network/deyes/types"

type Dispatcher struct {
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

func (d *Dispatcher) Start() {
}

func (d *Dispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	return nil
}

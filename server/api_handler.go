package server

import (
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
)

type ApiHandler struct {
	txProcessor *chains.TxProcessor
}

func NewApi(txProcessor *chains.TxProcessor) *ApiHandler {
	return &ApiHandler{
		txProcessor: txProcessor,
	}
}

// Empty function for checking health only.
func (api *ApiHandler) CheckHealth() {
}

// Called by Sisu to indicate that the server is ready to receive messages.
func (api *ApiHandler) SetSisuReady(chain string) {
}

// Adds a list of address to watch on a specific chain.
func (api *ApiHandler) AddWatchAddresses(chain string, addrs []string) {
	api.txProcessor.AddWatchAddresses(chain, addrs)
}

func (api *ApiHandler) DispatchTx(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	return api.txProcessor.DispatchTx(request)
}

func (api *ApiHandler) GetNonce(chain string, address string) int64 {
	return api.txProcessor.GetNonce(chain, address)
}

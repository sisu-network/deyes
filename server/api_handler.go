package server

import (
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/utils"
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

func (api *ApiHandler) DispatchTx(chain string, tx []byte) {
	// Dispatching a tx can take time.
	go func() {
		if err := api.txProcessor.DispatchTx(chain, tx); err != nil {
			utils.LogError("cannot dispatch tx, err =", err)
		}
	}()
}

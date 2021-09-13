package server

import "github.com/sisu-network/deyes/chains"

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

func (api *ApiHandler) DispatchTx(chain string, tx []byte, signature []byte) {

}

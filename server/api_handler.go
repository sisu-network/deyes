package server

import (
	ethcore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/core"
	"github.com/sisu-network/deyes/types"

	libchain "github.com/sisu-network/lib/chain"
)

type ApiHandler struct {
	processor *core.Processor
}

func NewApi(processor *core.Processor) *ApiHandler {
	return &ApiHandler{
		processor: processor,
	}
}

// Empty function for checking health only.
func (api *ApiHandler) Ping(source string) error {
	return nil
}

// Called by Sisu to indicate that the server is ready to receive messages.
func (api *ApiHandler) SetSisuReady(isReady bool) {
	api.processor.SetSisuReady(isReady)
}

func (api *ApiHandler) SetGatewayAddress(chain string, addr string) {
	api.processor.SetGateway(chain, addr)
}

func (api *ApiHandler) DispatchTx(request *types.DispatchedTxRequest) {
	api.processor.DispatchTx(request)
}

func (api *ApiHandler) GetNonce(chain string, address string) int64 {
	return api.processor.GetNonce(chain, address)
}

func (api *ApiHandler) GetGasPrices(chains []string) []int64 {
	prices := make([]int64, len(chains))
	for i, chain := range chains {
		if libchain.IsETHBasedChain(chain) {
			watcher := api.processor.GetWatcher(chain).(*ethcore.Watcher)
			prices[i] = watcher.GetGasPrice()
		} else {
			prices[i] = 0
		}
	}

	return prices
}

package server

import (
	"fmt"

	"github.com/echovl/cardano-go"
	chainscardano "github.com/sisu-network/deyes/chains/cardano"
	chainseth "github.com/sisu-network/deyes/chains/eth"
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

func (api *ApiHandler) SetVaultAddress(chain string, addr string) {
	api.processor.SetVault(chain, addr)
}

func (api *ApiHandler) DispatchTx(request *types.DispatchedTxRequest) {
	api.processor.DispatchTx(request)
}

func (api *ApiHandler) GetNonce(chain string, address string) int64 {
	return api.processor.GetNonce(chain, address)
}

// This API only applies for ETH chains.
func (api *ApiHandler) GetGasPrices(chains []string) []int64 {
	prices := make([]int64, len(chains))
	for i, chain := range chains {
		if libchain.IsETHBasedChain(chain) {
			watcher := api.processor.GetWatcher(chain).(*chainseth.Watcher)
			prices[i] = watcher.GetGasPrice()
		} else {
			prices[i] = 0
		}
	}

	return prices
}

func (api *ApiHandler) CardanoProtocolParams(chain string) (*cardano.ProtocolParams, error) {
	if !libchain.IsCardanoChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	return watcher.ProtocolParams()
}

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

type CardanoUtxosResult struct {
	Utxos []cardano.UTxO
	Bytes [][]byte
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

///// Carnado

func (api *ApiHandler) CardanoProtocolParams(chain string) (*cardano.ProtocolParams, error) {
	if !libchain.IsCardanoChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	return watcher.ProtocolParams()
}

func (api *ApiHandler) CardanoUtxos(chain string, addr string, maxBlock uint64) (CardanoUtxosResult, error) {
	if !libchain.IsCardanoChain(chain) {
		return CardanoUtxosResult{}, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	utxos, err := watcher.CardanoUtxos(addr, maxBlock)
	if err != nil {
		return CardanoUtxosResult{}, err
	}

	result := CardanoUtxosResult{
		Utxos: utxos,
	}
	// We have to marshal Amount since it's not serializable through network.
	result.Bytes = make([][]byte, len(utxos))
	for i, utxo := range utxos {
		result.Bytes[i], err = utxo.Amount.MarshalCBOR()
		if err != nil {
			return CardanoUtxosResult{}, err
		}
	}

	return result, err
}

// Balance returns the current balance of an account.
func (api *ApiHandler) CardanoBalance(chain string, address string, maxBlock int64) (*cardano.Value, error) {
	if !libchain.IsCardanoChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	return watcher.Balance(address, maxBlock)
}

// Tip returns the node's current tip
func (api *ApiHandler) CardanoTip(chain string) (*cardano.NodeTip, error) {
	return nil, nil
}

func (api *ApiHandler) CardanoSubmitTx(chain string, tx *cardano.Tx) (*cardano.Hash32, error) {

	return nil, nil
}

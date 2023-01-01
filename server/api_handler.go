package server

import (
	"fmt"

	"github.com/echovl/cardano-go"
	chainscardano "github.com/sisu-network/deyes/chains/cardano"
	chainseth "github.com/sisu-network/deyes/chains/eth"
	deyesethtypes "github.com/sisu-network/deyes/chains/eth/types"
	chainssolana "github.com/sisu-network/deyes/chains/solana"
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

func (api *ApiHandler) SetVaultAddress(chain string, addr string, token string) {
	api.processor.SetVault(chain, addr, token)
}

func (api *ApiHandler) DispatchTx(request *types.DispatchedTxRequest) {
	api.processor.DispatchTx(request)
}

func (api *ApiHandler) GetNonce(chain string, address string) (int64, error) {
	return api.processor.GetNonce(chain, address)
}

// This API only applies for ETH chains.
func (api *ApiHandler) GetGasInfo(chain string) *deyesethtypes.GasInfo {
	if libchain.IsETHBasedChain(chain) {
		watcher := api.processor.GetWatcher(chain).(*chainseth.Watcher)
		gasInfo := watcher.GetGasInfo()
		return &gasInfo
	} else {
		return nil
	}
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
func (api *ApiHandler) CardanoTip(chain string, blockHeight uint64) (*cardano.NodeTip, error) {
	if !libchain.IsCardanoChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	tip, err := watcher.Tip(blockHeight)
	return tip, err
}

func (api *ApiHandler) CardanoSubmitTx(chain string, tx *cardano.Tx) (*cardano.Hash32, error) {
	if !libchain.IsCardanoChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainscardano.Watcher)

	return watcher.SubmitTx(tx)
}

///// Solana
func (api *ApiHandler) SolanaQueryRecentBlock(chain string) (*types.SolanaQueryRecentBlockResult, error) {
	if !libchain.IsSolanaChain(chain) {
		return nil, fmt.Errorf("Invalid Cardano chain %s", chain)
	}

	watcher := api.processor.GetWatcher(chain).(*chainssolana.Watcher)
	hash, height, err := watcher.QueryRecentBlock()
	if err != nil {
		return nil, err
	}

	return &types.SolanaQueryRecentBlockResult{
		Hash:   hash,
		Height: height,
	}, nil
}

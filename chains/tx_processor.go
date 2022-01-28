package chains

import (
	"fmt"

	ethCore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"
)

// This struct handles the logic in deyes.
// TODO: Make this processor to support multiple chains at the same time.
type TxProcessor struct {
	db         database.Database
	txsCh      chan *types.Txs
	gasPriceCh chan *types.GasPriceRequest
	chain      string
	blockTime  int
	sisuClient client.Client

	watchers    map[string]Watcher
	dispatchers map[string]Dispatcher
	cfg         config.Deyes
}

func NewTxProcessor(cfg *config.Deyes, db database.Database, sisuClient client.Client) *TxProcessor {
	return &TxProcessor{
		cfg:         *cfg,
		db:          db,
		watchers:    make(map[string]Watcher),
		dispatchers: make(map[string]Dispatcher),
		sisuClient:  sisuClient,
	}
}

func (tp *TxProcessor) Start() {
	log.Info("Starting tx processor...")
	log.Info("tp.cfg.Chains = ", tp.cfg.Chains)

	tp.txsCh = make(chan *types.Txs, 1000)
	tp.gasPriceCh = make(chan *types.GasPriceRequest, 1000)

	for chain, cfg := range tp.cfg.Chains {
		go tp.listen()

		log.Info("Supported chain and config: ", chain, cfg)

		if libchain.IsETHBasedChain(chain) {
			watcher := ethCore.NewWatcher(tp.db, cfg, tp.txsCh, tp.gasPriceCh)
			tp.watchers[chain] = watcher
			go watcher.Start()

			// Dispatcher
			dispatcher := NewEhtDispatcher(chain, cfg.RpcUrl)
			dispatcher.Start()
			tp.dispatchers[chain] = dispatcher
		} else {
			panic(fmt.Errorf("Unknown chain %s", chain))
		}
	}
}

func (tp *TxProcessor) listen() {
	for {
		select {
		case txs := <-tp.txsCh:
			tp.sisuClient.BroadcastTxs(txs)
		case gasReq := <-tp.gasPriceCh:
			tp.sisuClient.UpdateGasPrice(gasReq)
		}
	}
}

func (tp *TxProcessor) AddWatchAddresses(chain string, addrs []string) {
	watcher := tp.watchers[chain]
	if watcher != nil {
		for _, addr := range addrs {
			log.Info("Adding watched addr ", addr, " for chain ", chain)
			watcher.AddWatchAddr(addr)
		}
	}
}

func (tp *TxProcessor) DispatchTx(request *types.DispatchedTxRequest) {
	chain := request.Chain

	dispatcher := tp.dispatchers[chain]
	if dispatcher == nil {
		types.NewDispatchTxError(fmt.Errorf("unknown chain %s", chain))
	}

	result := dispatcher.Dispatch(request)
	log.Info("Posting result to sisu for chain ", chain, " tx hash = ", request.TxHash)
	tp.sisuClient.PostDeploymentResult(result)
}

func (tp *TxProcessor) GetNonce(chain string, address string) int64 {
	watcher := tp.watchers[chain]
	if watcher == nil {
		return -1
	}

	return watcher.GetNonce(address)
}

func (tp *TxProcessor) GetWatcher(chain string) Watcher {
	return tp.watchers[chain]
}

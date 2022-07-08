package core

import (
	"fmt"
	"sync/atomic"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/chains"
	carcore "github.com/sisu-network/deyes/chains/cardano/core"
	"github.com/sisu-network/deyes/chains/eth-family/core"
	ethcore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/sisu-network/lib/log"

	"github.com/sisu-network/deyes/core/oracle"
)

// This struct handles the logic in deyes.
// TODO: Make this processor to support multiple chains at the same time.
type Processor struct {
	db            database.Database
	txsCh         chan *types.Txs
	priceUpdateCh chan []*types.TokenPrice
	chain         string
	blockTime     int
	sisuClient    client.Client

	watchers    map[string]chains.Watcher
	dispatchers map[string]chains.Dispatcher
	cfg         config.Deyes
	tpm         oracle.TokenPriceManager

	sisuReady atomic.Value
}

func NewProcessor(
	cfg *config.Deyes,
	db database.Database,
	sisuClient client.Client,
	tpm oracle.TokenPriceManager,
) *Processor {
	return &Processor{
		cfg:         *cfg,
		db:          db,
		watchers:    make(map[string]chains.Watcher),
		dispatchers: make(map[string]chains.Dispatcher),
		sisuClient:  sisuClient,
		tpm:         tpm,
	}
}

func (p *Processor) Start() {
	log.Info("Starting tx processor...")
	log.Info("tp.cfg.Chains = ", p.cfg.Chains)

	p.txsCh = make(chan *types.Txs, 1000)
	p.priceUpdateCh = make(chan []*types.TokenPrice)

	go p.listen()
	go p.tpm.Start(p.priceUpdateCh)

	for chain, cfg := range p.cfg.Chains {
		log.Info("Supported chain and config: ", chain, cfg)

		var watcher chains.Watcher
		var dispatcher chains.Dispatcher
		if libchain.IsETHBasedChain(chain) { // ETH chain
			watcher = ethcore.NewWatcher(p.db, cfg, p.txsCh, p.getEthClients(cfg.Rpcs))
			dispatcher = ethcore.NewEhtDispatcher(chain, cfg.Rpcs)
		} else if libchain.IsCardanoChain(chain) { // Cardano chain
			client := carcore.NewBlockfrostClient(
				blockfrost.APIClientOptions{
					ProjectID: cfg.RpcSecret,
					Server:    blockfrost.CardanoTestNet,
				},
			)

			watcher = carcore.NewWatcher(cfg, p.db, p.txsCh, client)
			dispatcher = carcore.NewDispatcher(client)
		} else {
			panic(fmt.Errorf("Unknown chain %s", chain))
		}

		p.watchers[chain] = watcher
		go watcher.Start()
		p.dispatchers[chain] = dispatcher
		dispatcher.Start()
	}
}

func (p *Processor) getEthClients(rpcs []string) []ethcore.EthClient {
	clients := core.NewEthClients(rpcs)
	if len(clients) == 0 {
		panic(fmt.Sprintf("None of the rpc server works, rpcs = %v", rpcs))
	}

	return clients
}

func (p *Processor) listen() {
	for {
		select {
		case txs := <-p.txsCh:
			if p.sisuReady.Load() == true {
				p.sisuClient.BroadcastTxs(txs)
			} else {
				log.Warnf("txs: Sisu is not ready")
			}
		case prices := <-p.priceUpdateCh:
			log.Infof("There is new token price update: %+v\n", prices)
			if p.sisuReady.Load() == true {
				p.sisuClient.UpdateTokenPrices(prices)
			} else {
				log.Warnf("prices: Sisu is not ready")
			}
		}
	}
}

func (tp *Processor) AddWatchAddresses(chain string, addrs []string) {
	log.Verbose("Received watch address from sisu: ", chain, addrs)

	watcher := tp.watchers[chain]
	if watcher != nil {
		for _, addr := range addrs {
			log.Info("Adding watched addr ", addr, " for chain ", chain)
			watcher.AddWatchAddr(addr)
		}
	} else {
		log.Critical("Watcher is nil")
	}
}

func (tp *Processor) DispatchTx(request *types.DispatchedTxRequest) {
	chain := request.Chain

	dispatcher := tp.dispatchers[chain]
	var result *types.DispatchedTxResult
	if dispatcher == nil {
		result = types.NewDispatchTxError(fmt.Errorf("unknown chain %s", chain))
	} else {
		result = dispatcher.Dispatch(request)
	}

	log.Info("Posting result to sisu for chain ", chain, " tx hash = ", request.TxHash, " success = ", result.Success)
	tp.sisuClient.PostDeploymentResult(result)
}

func (tp *Processor) GetNonce(chain string, address string) int64 {
	if !libchain.IsETHBasedChain(chain) {
		return -1
	}

	watcher := tp.watchers[chain].(*ethcore.Watcher)
	if watcher == nil {
		return -1
	}

	return watcher.GetNonce(address)
}

func (tp *Processor) GetWatcher(chain string) chains.Watcher {
	return tp.watchers[chain]
}

func (p *Processor) SetSisuReady(isReady bool) {
	p.sisuReady.Store(isReady)
}

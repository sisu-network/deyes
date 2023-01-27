package core

import (
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/chains/cardano"
	chainseth "github.com/sisu-network/deyes/chains/eth"
	chainlisk "github.com/sisu-network/deyes/chains/lisk"

	"github.com/sisu-network/deyes/chains/solana"
	chainstypes "github.com/sisu-network/deyes/chains/types"
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
	db         database.Database
	txsCh      chan *types.Txs
	txTrackCh  chan *chainstypes.TrackUpdate
	chain      string
	blockTime  int
	sisuClient client.Client

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
	p.txTrackCh = make(chan *chainstypes.TrackUpdate, 1000)

	go p.listen()
	go p.tpm.Start()

	for chain, cfg := range p.cfg.Chains {
		log.Info("Supported chain and config: ", chain, cfg)

		var watcher chains.Watcher
		var dispatcher chains.Dispatcher
		if libchain.IsETHBasedChain(chain) {
			// ETH chain
			client := chainseth.NewEthClients(cfg, p.cfg.UseExternalRpcsInfo)
			client.Start()

			watcher = chainseth.NewWatcher(p.db, cfg, p.txsCh, p.txTrackCh, client)
			dispatcher = chainseth.NewEhtDispatcher(chain, client)
		} else if libchain.IsCardanoChain(chain) {
			// Cardano chain
			client := p.getCardanoClient(cfg)
			watcher = cardano.NewWatcher(cfg, p.db, p.txsCh, p.txTrackCh, client)
			dispatcher = cardano.NewDispatcher(client)

		} else if libchain.IsSolanaChain(chain) {
			// Solana
			watcher = solana.NewWatcher(cfg, p.db, p.txsCh, p.txTrackCh)
			dispatcher = solana.NewDispatcher(cfg.Rpcs, cfg.Wss)

		} else if libchain.IsLiskChain(chain) {
			client := chainlisk.NewLiskClient(cfg)
			watcher = chainlisk.NewWatcher(p.db, cfg, p.txsCh, p.txTrackCh, client)
			dispatcher = chainlisk.NewDispatcher(chain, client)

		} else {
			panic(fmt.Errorf("Unknown chain %s", chain))
		}

		p.watchers[chain] = watcher
		go watcher.Start()

		p.dispatchers[chain] = dispatcher
		dispatcher.Start()
	}
}

func (p *Processor) getCardanoClient(cfg config.Chain) *cardano.DefaultCardanoClient {
	var (
		provider  cardano.Provider
		submitURL string
	)

	if cfg.ClientType == config.ClientTypeBlockFrost && len(cfg.RpcSecret) > 0 {
		log.Info("Use blockfrost API client")
		// TODO: Make this configurable
		provider = cardano.NewBlockfrostProvider(cfg)
		submitURL = "https://cardano-preprod.blockfrost.io/api/v0" + "/tx/submit"
	} else if cfg.ClientType == config.ClientTypeSelfHost {
		log.Info("Use Self-host client")
		db, err := cardano.ConnectDB(cfg.SyncDB)
		if err != nil {
			panic(err)
		}

		provider = cardano.NewSyncDBConnector(db)
		submitURL = cfg.SyncDB.SubmitURL
	} else {
		panic(fmt.Errorf("unknown cardano client type: %s", cfg.ClientType))
	}

	return cardano.NewDefaultCardanoClient(
		provider,
		submitURL,
		cfg.RpcSecret, // only used for Blockfrost API
	)
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

		case txTrackUpdate := <-p.txTrackCh:
			log.Verbose("There is a tx to confirm with hash: ", txTrackUpdate.Hash)
			p.sisuClient.OnTxIncludedInBlock(txTrackUpdate)
		}
	}
}

func (tp *Processor) SetVault(chain, addr string, token string) {
	log.Infof("Setting gateway, chain = %s, addr = %s", chain, addr)
	watcher := tp.GetWatcher(chain)
	watcher.SetVault(addr, token)
}

func (tp *Processor) DispatchTx(request *types.DispatchedTxRequest) {
	chain := request.Chain
	watcher := tp.GetWatcher(chain)
	if watcher == nil {
		log.Errorf("Cannot find watcher for chain %s", chain)
		tp.sisuClient.PostDeploymentResult(types.NewDispatchTxError(request, types.ErrGeneric))
		return
	}

	// If dispatching successful, add the tx to tracking.
	watcher.TrackTx(request.TxHash)

	dispatcher := tp.dispatchers[chain]
	var result *types.DispatchedTxResult
	if dispatcher == nil {
		log.Error(fmt.Errorf("Cannot find dispatcher for chain %s", chain))
		result = types.NewDispatchTxError(request, types.ErrGeneric)
	} else {
		log.Verbosef("Dispatching tx for chain %s with hash %s", request.Chain, request.TxHash)
		result = dispatcher.Dispatch(request)
	}

	log.Info("Posting result to sisu for chain ", chain, " tx hash = ", request.TxHash, " success = ", result.Success)
	tp.sisuClient.PostDeploymentResult(result)
}

func (tp *Processor) GetNonce(chain string, address string) (int64, error) {
	watcher := tp.GetWatcher(chain)
	if watcher == nil {
		return 0, fmt.Errorf("Cannot find watcher for chain %s", chain)
	}

	switch {
	case libchain.IsETHBasedChain(chain):
		return watcher.(*chainseth.Watcher).GetNonce(address)

	case libchain.IsLiskChain(chain):
		return watcher.(*chainlisk.Watcher).GetNonce(address)

	default:
		return 0, fmt.Errorf("unsupported chain for getting nonce, chain = %s", chain)
	}
}

func (tp *Processor) GetWatcher(chain string) chains.Watcher {
	return tp.watchers[chain]
}

func (p *Processor) SetSisuReady(isReady bool) {
	p.sisuReady.Store(isReady)
}

func (tp *Processor) GetTokenPrice(id string) (*big.Int, error) {
	return tp.tpm.GetTokenPrice(id)
}

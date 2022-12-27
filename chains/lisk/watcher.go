package lisk

import (
	"fmt"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sisu-network/deyes/chains"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"strings"
	"sync"
)

type Watcher struct {
	cfg       config.Chain
	clients   []LiskClient
	blockTime int
	db        database.Database
	txsCh     chan *types.Txs
	txTrackCh chan *chainstypes.TrackUpdate
	vault     string
	lock      *sync.RWMutex
	blockCh   chan *etypes.Block
}

func (w *Watcher) SetVault(addr string, token string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Verbosef("Setting vault for chain %s with address %s", w.cfg.Chain, addr)
	err := w.db.SetVault(w.cfg.Chain, addr, token)
	if err == nil {
		w.vault = strings.ToLower(addr)
	} else {
		log.Error("Failed to save gateway")
	}
}

func (w *Watcher) TrackTx(txHash string) {
}

func NewWatcher(db database.Database, cfg config.Chain,
	txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, clients []LiskClient) chains.Watcher {
	blockCh := make(chan *etypes.Block)

	w := &Watcher{
		db:        db,
		cfg:       cfg,
		clients:   clients,
		blockCh:   blockCh,
		txsCh:     txsCh,
		txTrackCh: txTrackCh,
	}
	return w
}

func (w *Watcher) init() {
	vaults, err := w.db.GetVaults(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	if len(vaults) > 0 {
		w.vault = vaults[0]
		log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.vault)
	} else {
		log.Infof("Vault for chain %s is not set yet", w.cfg.Chain)
	}
}

func (w *Watcher) Start() {
	log.Infof("Starting Watcher...")

	w.init()

	go w.waitForReceipt()
}

// waitForReceipt waits for receipts returned by the socket.
func (w *Watcher) waitForReceipt() {
	for _, client := range w.clients {
		for {
			tx := <-client.GetTransaction()
			fmt.Println(tx)
		}
		client.UpdateSocket()
	}

}

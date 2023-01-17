package lisk

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	ctypes "github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

const (
	TxTrackCacheSize = 1_000
)

type Watcher struct {
	cfg          config.Chain
	client       Client
	blockTime    int
	db           database.Database
	vault        string
	txTrackCache *lru.Cache
	lock         *sync.RWMutex
	txsCh        chan *ctypes.Txs
	doneCh       chan bool

	// Block fetcher
	blockCh      chan *types.Block
	blockFetcher BlockFetcher
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *ctypes.Txs,
	client Client) chains.Watcher {
	blockCh := make(chan *types.Block)

	w := &Watcher{
		blockCh:      blockCh,
		blockFetcher: newBlockFetcher(cfg, blockCh, client),
		db:           db,
		cfg:          cfg,
		txsCh:        txsCh,
		blockTime:    cfg.BlockTime,
		client:       client,
		lock:         &sync.RWMutex{},
		txTrackCache: lru.New(TxTrackCacheSize),
		doneCh:       make(chan bool),
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

func (w *Watcher) SetVault(addr string, token string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Verbosef("Setting vault for chain %s with address %s", w.cfg.Chain, addr)
	err := w.db.SetVault(w.cfg.Chain, addr, token)
	if err == nil {
		w.vault = strings.ToLower(addr)
	} else {
		log.Error("Failed to save vault")
	}
}

func (w *Watcher) Start() {
	log.Infof("Starting Watcher for chain %s", w.cfg.Chain)
	w.init()
	go w.scanBlocks()
}

func (w *Watcher) Stop() {
	w.blockFetcher.stop()
	w.doneCh <- true
}

func (w *Watcher) scanBlocks() {
	go w.blockFetcher.start()
	go w.waitForBlock()
}

// waitForBlock waits for new blocks from the block fetcher. It then filters interested txs and
func (w *Watcher) waitForBlock() {
	for {
		select {
		case <-w.doneCh:
			return

		case block := <-w.blockCh:
			// Pass this block to the receipt fetcher
			log.Info(w.cfg.Chain, " Block length = ", block.NumberOfTransactions)
			txArr := make([]*ctypes.Tx, 0)

			for _, tx := range block.Transactions {
				if tx.Sender != nil && strings.EqualFold(tx.Asset.Recipient.Address, w.vault) {
					if err := tx.Validate(); err != nil {
						log.Errorf("Failed to validate transaction, err = %v", err)
						continue
					}

					log.Infof("Transfer %s from address %s to %s with message %s",
						tx.Asset.Amount,
						tx.Sender.Address,
						tx.Asset.Recipient.Address,
						tx.Asset.Data,
					)
					bz, err := json.Marshal(tx)
					if err != nil {
						log.Errorf("Failed to marshal transaction, err = %v", err)
						continue
					}

					txFormatted := ctypes.Tx{
						Hash:       tx.Id,
						Serialized: bz,
						From:       tx.Sender.Address,
						To:         tx.Asset.Recipient.Address,
						Success:    tx.IsPending == false,
					}
					txArr = append(txArr, &txFormatted)
				}
			}

			if len(txArr) > 0 {
				txs := ctypes.Txs{
					Chain:     w.cfg.Chain,
					Block:     int64(block.Height),
					BlockHash: block.Id,
					Arr:       txArr,
				}
				w.txsCh <- &txs
			}
		}
	}
}

func (w *Watcher) processBlock(block *types.Block) []*types.Transaction {
	ret := make([]*types.Transaction, 0)
	for _, tx := range block.Transactions {
		if _, ok := w.txTrackCache.Get(tx.Id); ok {
			ret = append(ret, tx)
			continue
		}

		if w.acceptTx(tx) {
			ret = append(ret, tx)
		}
	}

	return ret
}

func (w *Watcher) acceptTx(tx *types.Transaction) bool {
	if tx.Asset.Recipient.Address != "" {
		if strings.EqualFold(tx.Asset.Recipient.Address, w.vault) {
			return true
		}
	}

	return false
}

func (w *Watcher) TrackTx(txHash string) {
	log.Verbose("Tracking tx: ", txHash)
	w.txTrackCache.Add(txHash, true)
}

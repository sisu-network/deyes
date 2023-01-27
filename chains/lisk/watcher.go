package lisk

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/chains"
	lisktypes "github.com/sisu-network/deyes/chains/lisk/types"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	types "github.com/sisu-network/deyes/types"
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
	txsCh        chan *types.Txs
	txTrackCh    chan *chainstypes.TrackUpdate
	doneCh       chan bool

	// Block fetcher
	blockCh      chan *lisktypes.Block
	blockFetcher BlockFetcher
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, client Client) chains.Watcher {
	blockCh := make(chan *lisktypes.Block)

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
		txTrackCh:    txTrackCh,
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

			w.processBlock(block)
		}
	}
}

func (w *Watcher) GetNonce(address string) (int64, error) {
	acc, err := w.client.GetAccount(address)
	if err != nil {
		return 0, err
	}

	nonce, err := strconv.ParseInt(acc.Sequence.Nonce, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("cannot parse nonce, address string = %s", address)
	}

	return nonce, nil
}

func (w *Watcher) processBlock(block *lisktypes.Block) {
	txArr := make([]*types.Tx, 0)

	for _, tx := range block.Transactions {
		if _, ok := w.txTrackCache.Get(tx.Id); ok {
			log.Verbose("Confirming lisk tx with hash = ", tx.Id)

			result := chainstypes.TrackResultConfirmed
			// This is a transaction that we are tracking. Inform Sisu about this.
			w.txTrackCh <- &chainstypes.TrackUpdate{
				Chain:       w.cfg.Chain,
				Hash:        tx.Id,
				BlockHeight: int64(block.Height),
				Result:      result,
			}

			continue
		}

		if tx.Sender != nil && tx.Asset != nil && tx.Asset.Recipient != nil &&
			strings.EqualFold(tx.Asset.Recipient.Address, w.vault) {
			if err := tx.Validate(); err != nil {
				log.Errorf("Failed to validate transaction, err = %v", err)
				continue
			}

			log.Infof("There is a transaction of amount %s from address %s to %s with message %s",
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

			txFormatted := types.Tx{
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
		txs := types.Txs{
			Chain:     w.cfg.Chain,
			Block:     int64(block.Height),
			BlockHash: block.Id,
			Arr:       txArr,
		}
		w.txsCh <- &txs
	}
}

func (w *Watcher) TrackTx(txHash string) {
	log.Verbose("Tracking tx: ", txHash)
	w.txTrackCache.Add(txHash, true)
}

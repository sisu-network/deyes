package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blockfrost/blockfrost-go"
	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/chains/common"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"
)

var (
	BlockNotFound = fmt.Errorf("Block not found")
)

type Watcher struct {
	cfg             config.Chain
	db              database.Database
	txsCh           chan *types.Txs
	client          CardanoClient
	blockTime       int
	lastBlockHeight atomic.Int32
	gateway         string

	txTrackCh        chan *types.TrackUpdate
	lock             *sync.RWMutex
	txTrackCache     *lru.Cache
	blockTimeTracker *common.BlockTimeTracker
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs,
	txTrackCh chan *types.TrackUpdate, client CardanoClient) *Watcher {
	return &Watcher{
		cfg:              cfg,
		db:               db,
		txsCh:            txsCh,
		blockTime:        cfg.BlockTime,
		txTrackCh:        txTrackCh,
		lock:             &sync.RWMutex{},
		client:           client,
		txTrackCache:     lru.New(1000),
		blockTimeTracker: common.NewBlockTimeTracker(cfg.BlockTime),
	}
}

func (w *Watcher) init() {
	healthy := w.client.IsHealthy()
	if !healthy {
		err := fmt.Errorf("Blockfrost is not healthy")
		log.Error(err)
		panic(err)
	}

	var err error
	w.gateway, err = w.db.GetGateway(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	log.Infof("Saved gateway in the db for chain %s is %s", w.cfg.Chain, w.gateway)

	w.lastBlockHeight.Store(0)
}

func (w *Watcher) Start() {
	w.init()

	go w.scanChain()
}

func (w *Watcher) scanChain() {
	log.Info("Start scanning chain: ", w.cfg.Chain)

	for {
		w.blockTime = w.blockTimeTracker.GetSleepTime()
		log.Verbose("Block time on chain ", w.cfg.Chain, " is ", w.blockTime)

		// Get next block to scan
		block, err := w.getNextBlock()

		if err != nil {
			latestBlock, err2 := w.client.LatestBlock()
			if err2 != nil {
				continue
			} else {
				if !strings.Contains(err.Error(), "Not Found") {
					// This error is not a Block not found error. Print the error here.
					log.Errorf("Error when getting block for height/hash = %d, error = %v\n", latestBlock.Height, err)
				} else {

					log.Verbosef("Block %d not found. We need to wait more. w.blockTime = %d\n",
						latestBlock.Height, w.blockTime)
				}
			}

			w.blockTimeTracker.MissBlock()
			time.Sleep(time.Duration(w.blockTimeTracker.GetSleepTime()) * time.Millisecond)
			continue
		}

		// Process each address in the interested addr.
		txArr := make([]*types.Tx, 0)
		txsIn, err := w.client.NewTxs(block.Height, w.gateway)
		if err != nil {
			log.Error("Cannot get list of new transaction at block ", block.Height, " err = ", err)
			// This is a networking or RPC call error. We do not count this as a block miss.
			time.Sleep(time.Duration(w.blockTimeTracker.GetSleepTime()) * time.Millisecond)
			continue
		}

		log.Info("Block number on ", w.cfg.Chain, " = ", block.Height)

		w.lastBlockHeight.Store(int32(block.Height))
		w.blockTimeTracker.HitBlock()

		if len(w.gateway) == 0 {
			log.Verbose("Gateway is still empty")
			continue
		}

		log.Verbose("Filtered txs sizes = ", len(txsIn), " on chain ", w.cfg.Chain)

		for _, txIn := range txsIn {
			bz, err := json.Marshal(txIn)
			if err != nil {
				log.Error("Cannot serialize utxo, err = ", err)
				continue
			}

			if _, ok := w.txTrackCache.Get(txIn.Hash); ok {
				log.Verbose("Confirming cardano tx with hash = ", txIn.Hash)

				// This is a transction that we are tracking. Inform Sisu about this.
				w.txTrackCh <- &types.TrackUpdate{
					Chain:       w.cfg.Chain,
					Bytes:       bz,
					Hash:        txIn.Hash,
					BlockHeight: int64(block.Height),
					Result:      types.TrackResultConfirmed,
				}

				continue
			}

			txArr = append(txArr, &types.Tx{
				Hash:        utils.KeccakHash32(fmt.Sprintf("%s__%d", txIn.Hash, txIn.Index)),
				OutputIndex: txIn.Index,
				Serialized:  bz,
				To:          txIn.Address,
			})
		}

		if len(txArr) > 0 {
			txs := types.Txs{
				Chain:     w.cfg.Chain,
				Block:     int64(block.Height),
				BlockHash: block.Hash,
				Arr:       txArr,
			}

			// Broadcast the result
			w.txsCh <- &txs
		}

		// Sleep until next block
		time.Sleep(time.Duration(w.blockTimeTracker.GetSleepTime()) * time.Millisecond)
	}
}

func (w *Watcher) getNextBlock() (*blockfrost.Block, error) {
	lastScanBlock := int(w.lastBlockHeight.Load())
	nextBlock := lastScanBlock + 1
	if lastScanBlock == 0 {
		latestBlock, err := w.client.LatestBlock()
		if err != nil {
			return nil, err
		}

		nextBlock = latestBlock.Height
	}

	block, err := w.client.GetBlock(strconv.Itoa(nextBlock))
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (w *Watcher) SetChainAccount(addr string) {
	// Do nothing. In Cardano, the gateway is the same with chain account.
}

func (w *Watcher) SetGateway(addr string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetGateway(w.cfg.Chain, addr)
	if err == nil {
		w.gateway = addr
	} else {
		log.Error("Failed to save gateway")
	}
}

func (w *Watcher) TrackTx(txHash string) {
	fmt.Println("Cardano tracking tx: ", txHash)
	w.txTrackCache.Add(txHash, true)
}

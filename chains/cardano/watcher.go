package cardano

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	cardanogo "github.com/echovl/cardano-go"
	"github.com/golang/groupcache/lru"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"

	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
	chainstypes "github.com/sisu-network/deyes/chains/types"
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
	vault           string

	txTrackCh    chan *chainstypes.TrackUpdate
	lock         *sync.RWMutex
	txTrackCache *lru.Cache
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate, client CardanoClient) *Watcher {
	return &Watcher{
		cfg:          cfg,
		db:           db,
		txsCh:        txsCh,
		blockTime:    cfg.BlockTime,
		txTrackCh:    txTrackCh,
		lock:         &sync.RWMutex{},
		client:       client,
		txTrackCache: lru.New(1000),
	}
}

func (w *Watcher) init() {
	healthy := w.client.IsHealthy()
	if !healthy {
		err := fmt.Errorf("Blockfrost is not healthy")
		log.Error(err)
		panic(err)
	}

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

	w.lastBlockHeight.Store(0)
}

func (w *Watcher) Start() {
	w.init()

	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	log.Info("Start scanning chain: ", w.cfg.Chain)

	for {
		// Get next block to scan
		block, err := w.getNextBlock()

		if err != nil {
			w.blockTime = w.blockTime + w.cfg.AdjustTime
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			latestBlock, err2 := w.client.LatestBlock()
			if err2 != nil {
				log.Error(err2)
				continue
			}

			if !strings.Contains(err.Error(), "Not Found") {
				// This error is not a Block not found error. Print the error here.
				log.Errorf("Error when getting block for height/hash = %d, error = %v\n", latestBlock.Height, err)
			} else {
				log.Verbosef("%s: Block %d not found. We need to wait more. w.blockTime = %d\n",
					w.cfg.Chain, latestBlock.Height, w.blockTime)
			}
			continue
		}

		// Process each address in the interested addr.
		txArr := make([]*types.Tx, 0)
		txsIn, err := w.client.NewTxs(block.Height, w.vault)
		if err != nil {
			log.Error("Cannot get list of new transaction at block ", block.Height, " err = ", err)
			time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
			continue
		}

		log.Info("Block height && blocktime for Cardano Scanner = ", block.Height, w.blockTime)

		w.lastBlockHeight.Store(int32(block.Height))
		w.blockTime = w.blockTime - w.cfg.AdjustTime/4

		if len(w.vault) == 0 {
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
				w.txTrackCh <- &chainstypes.TrackUpdate{
					Chain:       w.cfg.Chain,
					Bytes:       bz,
					Hash:        txIn.Hash,
					BlockHeight: int64(block.Height),
					Result:      chainstypes.TrackResultConfirmed,
				}

				continue
			}

			txArr = append(txArr, &types.Tx{
				Hash:        utils.KeccakHash32(fmt.Sprintf("%s__%d", txIn.Hash, txIn.Index)),
				OutputIndex: txIn.Index,
				Serialized:  bz,
				To:          txIn.Address,
				Success:     true,
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
		time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
	}
}

func (w *Watcher) getNextBlock() (*providertypes.Block, error) {
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

func (w *Watcher) SetVault(addr string, token string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.db.SetVault(w.cfg.Chain, addr, "")
	if err == nil {
		w.vault = addr
	} else {
		log.Error("Failed to save gateway")
	}
}

func (w *Watcher) TrackTx(txHash string) {
	log.Verbosef("Tracking cardano tx with hash: %s", txHash)
	w.txTrackCache.Add(txHash, true)
}

func (w *Watcher) ProtocolParams() (*cardanogo.ProtocolParams, error) {
	return w.client.ProtocolParams()
}

func (w *Watcher) CardanoUtxos(addr string, maxBlock uint64) ([]cardanogo.UTxO, error) {
	return w.client.AddressUTXOs(context.Background(), addr, providertypes.APIQueryParams{
		To: fmt.Sprint(maxBlock),
	})
}

func (w *Watcher) Balance(address string, maxBlock int64) (*cardanogo.Value, error) {
	return w.client.Balance(address, maxBlock)
}

// Tip returns the node's current tip
func (w *Watcher) Tip(maxBlock uint64) (*cardanogo.NodeTip, error) {
	return w.client.Tip(maxBlock)
}

func (w *Watcher) SubmitTx(tx *cardanogo.Tx) (*cardanogo.Hash32, error) {
	return w.client.SubmitTx(tx)
}

package solana

import (
	"context"
	"fmt"
	"sync"
	"time"

	"encoding/json"

	"github.com/golang/groupcache/lru"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
	solanatypes "github.com/sisu-network/deyes/chains/solana/types"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"

	"github.com/ybbus/jsonrpc/v3"
)

const BLOCK_CHUNK_SIZE = 10

type Watcher struct {
	cfg          config.Chain
	lastSlot     atomic.Uint64
	client       jsonrpc.RPCClient
	txTrackCache *lru.Cache
	db           database.Database
	tokenAta     *sync.Map

	txsCh     chan *types.Txs
	txTrackCh chan *chainstypes.TrackUpdate
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate) *Watcher {

	return &Watcher{
		cfg:          cfg,
		db:           db,
		txsCh:        txsCh,
		txTrackCache: lru.New(1000),
		tokenAta:     &sync.Map{},
		txTrackCh:    txTrackCh,
		client:       jsonrpc.NewClient(cfg.Rpcs[0]),
	}
}

func (w *Watcher) Start() {
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	log.Verbose("Start scanning solana block...")

	for {
		time.Sleep(300 * time.Millisecond)

		block, err := w.getNextBlock()
		if err != nil || block == nil {
			continue
		} else {
			log.Verbose("Block height && blocktime for Solana Scanner = ", block.ParentSlot+1)
			if block.Transactions == nil {
				log.Error("Err is not nil but transactions list is nil. block = ", block)
				continue
			}

			txArr := make([]*types.Tx, 0)

			// Process all transaction in the block
			for _, outerTx := range block.Transactions {
				innerTx := outerTx.TransactionInner

				if len(innerTx.Signatures) == 0 {
					log.Error("Decoded solana transaction does not have signature")
					continue
				}

				bz, err := json.Marshal(outerTx)
				if err != nil {
					log.Error("Failed to marshal outer tx")
					continue
				}

				txId := innerTx.Signatures[0]

				if _, ok := w.txTrackCache.Get(txId); ok {
					log.Verbose("Confirming solana tx with hash = ", txId)

					result := chainstypes.TrackResultConfirmed
					if outerTx.Meta.Err != nil {
						result = chainstypes.TrackResultFailure
					}

					// This is a transction that we are tracking. Inform Sisu about this.
					w.txTrackCh <- &chainstypes.TrackUpdate{
						Chain:       w.cfg.Chain,
						Bytes:       bz,
						Hash:        txId,
						BlockHeight: int64(block.ParentSlot + 1),
						Result:      result,
					}

					continue
				}

				// Check to see if this is a transaction sent to one of our token accounts.
				if w.acceptTx(outerTx) {
					log.Verbose("Found a transaction sent to our bridge")

					txArr = append(txArr, &types.Tx{
						Hash:       outerTx.TransactionInner.Signatures[0],
						Serialized: bz,
						To:         w.cfg.SolanaBridgeProgramId,
						Success:    true,
					})
				}
			}
		}
	}
}

func (w *Watcher) acceptTx(outerTx *solanatypes.Transaction) bool {
	if outerTx.Meta.Err != nil {
		return false
	}

	if outerTx == nil || outerTx.TransactionInner == nil || outerTx.TransactionInner.Message == nil ||
		outerTx.TransactionInner.Message.AccountKeys == nil {
		return false
	}

	accounts := outerTx.TransactionInner.Message.AccountKeys
	programIdFound := false
	// Check that if the brdige program id is in the accounts
	for _, accountKey := range accounts {
		if accountKey == w.cfg.SolanaBridgeProgramId {
			programIdFound = true
			break
		}
	}

	if !programIdFound {
		return false
	}

	// Check that there is at least one instruction sent to the program id
	for _, ix := range outerTx.TransactionInner.Message.Instructions {
		if accounts[ix.ProgramIdIndex] == w.cfg.SolanaBridgeProgramId {
			return true
		}
	}

	return false
}

func (w *Watcher) getTransferOutData(outerTx *solanatypes.Transaction) (*solanatypes.TransferOutData, error) {
	accounts := outerTx.TransactionInner.Message.AccountKeys

	// Check that there is at least one instruction sent to the program id
	for _, ix := range outerTx.TransactionInner.Message.Instructions {
		if accounts[ix.ProgramIdIndex] != w.cfg.SolanaBridgeProgramId {
			continue
		}

		// Decode the instruction
		bytesArr, err := base58.Decode(ix.Data)
		if err != nil {
			return nil, err
		}

		if len(bytesArr) == 0 {
			return nil, fmt.Errorf("Data is empty")
		}

		borshBz := bytesArr[1:]
		transferData := new(solanatypes.TransferOutData)
		err = borsh.Deserialize(transferData, borshBz)
		if err != nil {
			return nil, err
		}

		return transferData, nil
	}

	return nil, fmt.Errorf("TransferOut Instruction not found")
}

func (w *Watcher) getNextBlock() (*solanatypes.Block, error) {
	var slot uint64
	slot = w.lastSlot.Load()
	if slot == 0 {
		var err error
		slot, err = w.getSlot()
		if err != nil {
			return nil, err
		}

		w.lastSlot.Store(slot - 1)
	}

	nextSlot := slot

	for {
		nextSlot = nextSlot + 1
		result, err := w.getBlockNumber(nextSlot)

		if err != nil {
			rpcErr, ok := err.(*jsonrpc.RPCError)
			if ok {
				// -32007: Slot 171913340 was skipped, or missing due to ledger jump to recent snapshot
				// -32015: Transaction version (0) is not supported by the requesting client. Please try the request again with the following configuration parameter: \"maxSupportedTransactionVersion\": 0"
				if rpcErr.Code == -32007 || rpcErr.Code == -32015 {
					// Slot is skipped, try the next one.
					w.lastSlot.Store(nextSlot)
					continue
				}
			}

			return nil, err
		}

		// Update last scanned block
		w.lastSlot.Store(nextSlot)

		return result, err
	}
}

func (w *Watcher) getSlot() (uint64, error) {
	res, err := w.client.Call(context.Background(), "getSlot")
	if err != nil {
		return 0, err
	}

	var result uint64
	err = res.GetObject(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (w *Watcher) getBlockNumber(slot uint64) (*solanatypes.Block, error) {
	var request = &solanatypes.GetBlockRequest{
		TransactionDetails:             "full",
		MaxSupportedTransactionVersion: 100,
	}

	log.Verbose("Getting block with slot ", slot)
	res, err := w.client.Call(context.Background(), "getBlock", slot, request)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		return nil, res.Error
	}

	block := new(solanatypes.Block)
	err = res.GetObject(&block)
	if err != nil {
		return nil, err
	}

	if block == nil {
		err := fmt.Errorf("Error is nil but block is also nil. Slot = %d", slot)
		log.Error(err)

		return nil, err
	}

	return block, nil
}

func (w *Watcher) SetVault(addr string, token string) {
	w.tokenAta.Store(addr, true)
}

func (w *Watcher) TrackTx(txHash string) {

}

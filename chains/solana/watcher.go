package solana

import (
	"context"
	"fmt"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/golang/groupcache/lru"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"
)

const BLOCK_CHUNK_SIZE = 10

type Watcher struct {
	cfg             config.Chain
	lastSlot        atomic.Uint64
	client          *rpc.Client
	txTrackCache    *lru.Cache
	db              database.Database
	lastBlockHeight atomic.Int32

	txsCh     chan *types.Txs
	txTrackCh chan *chainstypes.TrackUpdate
}

func NewWatcher(cfg config.Chain, db database.Database, txsCh chan *types.Txs,
	txTrackCh chan *chainstypes.TrackUpdate) *Watcher {
	var client *rpc.Client
	switch cfg.Chain {
	case "solana-devnet":
		client = rpc.New(rpc.DevNet_RPC)
		_ = client
	default:
		panic("Unsupported chain")
	}

	return &Watcher{
		cfg:          cfg,
		db:           db,
		txsCh:        txsCh,
		txTrackCache: lru.New(1000),
		txTrackCh:    txTrackCh,
		client:       client,
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
		if err != nil {
		} else {
			// Process all transaction in the block
			for _, outerTx := range block.Transactions {
				bz := outerTx.Transaction.GetBinary()
				decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(outerTx.Transaction.GetBinary()))
				if err != nil {
					log.Error("Failed to decode solana transaction")
					continue
				}

				if len(decodedTx.Signatures) == 0 {
					log.Error("Decoded solana transaction does not have signature")
					continue
				}

				txId := decodedTx.Signatures[0]
				if _, ok := w.txTrackCache.Get(txId.String()); ok {
					log.Verbose("Confirming solana tx with hash = ", txId.String())

					result := chainstypes.TrackResultConfirmed
					if outerTx.Meta.Err != nil {
						result = chainstypes.TrackResultFailure
					}

					// This is a transction that we are tracking. Inform Sisu about this.
					w.txTrackCh <- &chainstypes.TrackUpdate{
						Chain:       w.cfg.Chain,
						Bytes:       bz,
						Hash:        txId.String(),
						BlockHeight: int64(*block.BlockHeight),
						Result:      result,
					}

					continue
				}

				// Check to see if this is a transaction sent to one of our token accounts.
				if w.acceptTx(outerTx) {
				}
			}
		}
	}
}

func (w *Watcher) acceptTx(outerTx rpc.TransactionWithMeta) bool {
	return false
}

func (w *Watcher) getNextBlock() (*rpc.GetBlockResult, error) {
	var slot uint64
	slot = w.lastSlot.Load()
	if slot == 0 {
		var err error
		slot, err = w.client.GetSlot(context.Background(), rpc.CommitmentFinalized)
		if err != nil {
			return nil, err
		}

		w.lastSlot.Store(slot - 1)
	}

	nextSlot := slot

	for {
		nextSlot = nextSlot + 1
		fmt.Println("nextSlot = ", nextSlot)

		result, err := w.getBlockNumber(nextSlot)
		if err != nil {
			rpcErr, ok := err.(*jsonrpc.RPCError)

			fmt.Println("Error scanning block, err = ", err)
			if ok {
				fmt.Println("rpcErr.Code = ", rpcErr.Code)
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

func (w *Watcher) getBlockNumber(slot uint64) (*rpc.GetBlockResult, error) {
	maxTxVersion := uint64(10)

	return w.client.GetBlockWithOpts(context.Background(), slot, &rpc.GetBlockOpts{
		// TransactionDetails: "full",
		MaxSupportedTransactionVersion: &maxTxVersion,
	})
}

func (w *Watcher) SetVault(addr string, token string) {

}

func (w *Watcher) TrackTx(txHash string) {

}

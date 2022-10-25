package solana

import (
	"context"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	lru "github.com/hashicorp/golang-lru"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"go.uber.org/atomic"
)

const BLOCK_CHUNK_SIZE = 10

type Watcher struct {
	cfg          config.Chain
	lastSlot     atomic.Uint64
	client       *rpc.Client
	txTrackCache *lru.Cache
	db           database.Database

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
		cfg:       cfg,
		db:        db,
		txsCh:     txsCh,
		txTrackCh: txTrackCh,
		client:    client,
	}
}

func (w *Watcher) Start() {
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	for {
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

		w.lastSlot.Store(slot)
	}

	return w.getBlockNumber(slot)
}

func (w *Watcher) getBlockNumber(slot uint64) (*rpc.GetBlockResult, error) {
	return w.client.GetBlock(context.Background(), slot)
}

func (w *Watcher) SetVault(addr string, token string) {

}

func (w *Watcher) TrackTx(txHash string) {

}

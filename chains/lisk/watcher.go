package lisk

import (
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/lib/log"
)

type Watcher struct {
	cfg config.Chain
	db  database.Database
}

func (w *Watcher) SetVault(addr string, token string) {

}

func (w *Watcher) TrackTx(txHash string) {

}

func NewWatcher(db database.Database, cfg config.Chain) chains.Watcher {
	w := &Watcher{
		db:  db,
		cfg: cfg,
	}
	//w.
	return w
}

func (w *Watcher) init() {

}

func (w *Watcher) Start() {
	log.Info("Starting Watcher...")

	w.init()
	//go w.scanBlocks()
}

// extractTxs takes resposne from the receipt fetcher and converts them into deyes transactions.
//func (w *Watcher) extractTxs(response *txReceiptResponse) *types.Txs {
//	arr := make([]*types.Tx, 0)
//	for i, tx := range response.txs {
//		receipt := response.receipts[i]
//		bz, err := tx.MarshalBinary()
//		if err != nil {
//			log.Error("Cannot serialize ETH tx, err = ", err)
//			continue
//		}
//
//		if _, ok := w.txTrackCache.Get(tx.Hash().String()); ok {
//			// Get Tx Receipt
//			result := chainstypes.TrackResultConfirmed
//			if receipt.Status == 0 {
//				result = chainstypes.TrackResultFailure
//			}
//
//			// This is a transaction that we are tracking. Inform Sisu about this.
//			w.txTrackCh <- &chainstypes.TrackUpdate{
//				Chain:       w.cfg.Chain,
//				Bytes:       bz,
//				Hash:        tx.Hash().String(),
//				BlockHeight: response.blockNumber,
//				Result:      result,
//			}
//
//			continue
//		}
//
//		var to string
//		if tx.To() == nil {
//			to = ""
//		} else {
//			to = tx.To().String()
//		}
//
//		from, err := w.getFromAddress(w.cfg.Chain, tx)
//		if err != nil {
//			log.Errorf("cannot get from address for tx %s on chain %s, err = %v", tx.Hash().String(), w.cfg.Chain, err)
//			continue
//		}
//
//		arr = append(arr, &types.Tx{
//			Hash:       tx.Hash().String(),
//			Serialized: bz,
//			From:       from.Hex(),
//			To:         to,
//			Success:    receipt.Status == 1,
//		})
//	}
//
//	return &types.Txs{
//		Chain:     w.cfg.Chain,
//		Block:     response.blockNumber,
//		BlockHash: response.blockHash,
//		Arr:       arr,
//	}
//}

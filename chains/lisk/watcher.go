package lisk

import (
	"encoding/hex"
	"fmt"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
	"github.com/sisu-network/deyes/chains"
	lisk "github.com/sisu-network/deyes/chains/lisk/types"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
	"strings"
	"sync"
	"unsafe"
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
			txs := w.extractTxs(tx.Params.Transaction)
			log.Verbose(w.cfg.Chain, ": txs sizes = ", len(txs.Arr))

			if len(txs.Arr) > 0 {
				// Send list of interested txs back to the listener.
				w.txsCh <- txs
			}

			// Save all txs into database for later references.
			w.db.SaveTxs(w.cfg.Chain, 0, txs)
		}
		client.UpdateSocket()
	}

}

// extractTxs takes response from the receipt socket and converts them into deyes transactions.
func (w *Watcher) extractTxs(response string) *types.Txs {
	data, _ := hex.DecodeString(response)
	transaction := &lisk.TransactionMessage{}
	if err := proto.Unmarshal(data, transaction); err != nil {
		log.Errorf("Failed to parse  transaction:", err)
	}

	asset := &lisk.AssetMessage{}
	if err := proto.Unmarshal(transaction.Asset, asset); err != nil {
		log.Errorf("Failed to parse  asset:", err)
	}
	fmt.Println(asset)
	arr := make([]*types.Tx, 0)

	tx := &types.Tx{
		Serialized: transaction.Signatures[0],
		Success:    true,
		From:       hex.EncodeToString(transaction.SenderPublicKey),
		To:         hex.EncodeToString(asset.RecipientAddress),
	}
	arr = append(arr, tx)
	return &types.Txs{
		Chain: w.cfg.Chain,
		Block: int64(uintptr(unsafe.Pointer(&transaction.Nonce))),
		//BlockHash: response.blockHash,
		Arr: arr,
	}
}

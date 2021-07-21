package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

type Watcher struct {
	chain       string
	rpcEndpoint string
	client      *ethclient.Client
	blockHeight int64
	blockTime   int
	db          *database.Database
	txsCh       chan *types.Txs
	// A set of address we are interested in. Only send information about transaction to these
	// addresses back to Sisu.
	interestedAddrs *sync.Map
}

func NewWatcher(db *database.Database, rpcEndpoint string, blockTime int, chain string, txsCh chan *types.Txs) *Watcher {
	return &Watcher{
		db:              db,
		rpcEndpoint:     rpcEndpoint,
		blockTime:       blockTime,
		chain:           chain,
		txsCh:           txsCh,
		interestedAddrs: &sync.Map{},
	}
}

func (w *Watcher) init() {
	var err error
	w.client, err = ethclient.Dial(w.rpcEndpoint)
	if err != nil {
		panic(err)
	}

	blockHeight, err := w.db.LoadBlockHeight(w.chain)
	if err != nil {
		panic(err)
	}

	fmt.Println("blockHeight from db = ", blockHeight)

	w.blockHeight = blockHeight
}

func (w *Watcher) Start() {
	utils.LogInfo("Starting Watcher...")

	w.init()
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	for {
		// Get the blockheight
		block, err := w.getBlock(w.blockHeight)
		fmt.Println("err = ", err)

		switch err {
		case nil:
			fmt.Println("Height = ", block.Number())
		case ethereum.NotFound:
			// Ping block for every second.
		}

		allTxs, err := w.processBlock(block)
		if err == nil {
			// Filter only transactions that we are interested in.
			filtered := w.filterTxs(allTxs)

			if len(filtered.Arr) > 0 {
				// Send list of interested txs back to the listener.
				w.txsCh <- filtered
			}

			// Save all txs into database for later references.
			w.db.SaveTxs(w.chain, w.blockHeight, allTxs)

			w.blockHeight++
		}

		time.Sleep(time.Duration(w.blockTime) * time.Millisecond)
	}
}

func (w *Watcher) getBlock(height int64) (*etypes.Block, error) {
	return w.client.BlockByNumber(context.Background(), big.NewInt(height))
}

func (w *Watcher) processBlock(block *etypes.Block) (*types.Txs, error) {
	arr := make([]*types.Tx, 0)

	for _, tx := range block.Transactions() {
		bz, err := tx.MarshalJSON()
		if err != nil {
			utils.LogError("Cannot serialize ETH tx, err = ", err)
			continue
		}

		arr = append(arr, &types.Tx{
			Serialized: bz,
		})
	}

	return &types.Txs{
		Chain: w.chain,
		Arr:   arr,
	}, nil
}

func (w *Watcher) filterTxs(txs *types.Txs) *types.Txs {
	arr := make([]*types.Tx, 0)

	for _, tx := range txs.Arr {
		// Check if we are interested in this transaction.
		_, ok := w.interestedAddrs.Load(tx.To)
		if ok {
			arr = append(arr, tx)
		}
	}

	return &types.Txs{
		Chain: w.chain,
		Arr:   arr,
	}
}

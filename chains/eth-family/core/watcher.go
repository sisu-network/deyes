package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

// TODO: Move this to the chains package.
type Watcher struct {
	cfg         config.Chain
	client      *ethclient.Client
	blockHeight int64
	blockTime   int
	db          database.Database
	txsCh       chan *types.Txs
	// A set of address we are interested in. Only send information about transaction to these
	// addresses back to Sisu.
	interestedAddrs *sync.Map

	signers map[string]etypes.Signer
}

func NewWatcher(db database.Database, cfg config.Chain, txsCh chan *types.Txs) *Watcher {
	return &Watcher{
		db:              db,
		cfg:             cfg,
		txsCh:           txsCh,
		interestedAddrs: &sync.Map{},
	}
}

func (w *Watcher) init() {
	var err error

	utils.LogInfo("RPC endpoint for chain", w.cfg.Chain, "is", w.cfg.RpcUrl)
	w.client, err = ethclient.Dial(w.cfg.RpcUrl)
	if err != nil {
		panic(err)
	}

	blockHeight, err := w.db.LoadBlockHeight(w.cfg.Chain)
	if err != nil {
		panic(err)
	}

	utils.LogInfo("startingBlock = ", w.cfg.StartingBlock)

	w.blockHeight = utils.MaxInt(int64(w.cfg.StartingBlock), blockHeight)
}

func (w *Watcher) AddWatchAddr(addr string) {
	w.interestedAddrs.Store(addr, true)
}

func (w *Watcher) Start() {
	utils.LogInfo("Starting Watcher...")

	w.init()
	go w.scanBlocks()
}

func (w *Watcher) scanBlocks() {
	latestBlock, err := w.getLatestBlock()
	if err == nil {
		w.blockHeight = latestBlock.Header().Number.Int64()
	}
	utils.LogInfo("Latest height = ", w.blockHeight)

	for {
		// Get the blockheight
		block, err := w.tryGetBlock()
		if err != nil || block == nil {
			utils.LogError("Cannot get block at height", w.blockHeight, "for chain", w.cfg.Chain)
			time.Sleep(time.Duration(w.cfg.BlockTime) * time.Millisecond)
			continue
		}

		filteredTxs, err := w.processBlock(block)
		utils.LogVerbose("Filtered txs sizes = ", len(filteredTxs.Arr))

		if err == nil {
			if len(filteredTxs.Arr) > 0 {
				// Send list of interested txs back to the listener.
				w.txsCh <- filteredTxs
			}

			// Save all txs into database for later references.
			w.db.SaveTxs(w.cfg.Chain, w.blockHeight, filteredTxs)

			w.blockHeight++
		}

		time.Sleep(time.Duration(w.cfg.BlockTime) * time.Millisecond)
	}
}

// Get block with retry when block is not mined yet.
func (w *Watcher) tryGetBlock() (*etypes.Block, error) {
	block, err := w.getBlock(w.blockHeight)
	switch err {
	case nil:
		utils.LogDebug("Height = ", block.Number())
		return block, nil

	case ethereum.NotFound:
		// TODO: Ping block for every second.
		for i := 0; i < 10; i++ {
			block, err = w.getBlock(w.blockHeight)
			if err == nil {
				return block, err
			}

			time.Sleep(time.Duration(time.Second))
		}
	}

	return block, err
}

func (w *Watcher) getLatestBlock() (*etypes.Block, error) {
	return w.client.BlockByNumber(context.Background(), nil)
}

func (w *Watcher) getBlock(height int64) (*etypes.Block, error) {
	return w.client.BlockByNumber(context.Background(), big.NewInt(height))
}

func (w *Watcher) processBlock(block *etypes.Block) (*types.Txs, error) {
	arr := make([]*types.Tx, 0)

	utils.LogInfo("Block length = ", len(block.Transactions()))

	for _, tx := range block.Transactions() {
		bz, err := tx.MarshalBinary()
		if err != nil {
			utils.LogError("Cannot serialize ETH tx, err = ", err)
			continue
		}

		// Only filter our interested transaction.
		if !w.acceptTx(tx) {
			continue
		}

		var to string
		if tx.To() == nil {
			to = ""
		} else {
			to = tx.To().String()
		}

		arr = append(arr, &types.Tx{
			Hash:       tx.Hash().String(),
			Serialized: bz,
			To:         to,
		})
	}

	return &types.Txs{
		Chain: w.cfg.Chain,
		Arr:   arr,
	}, nil
}

func (w *Watcher) acceptTx(tx *etypes.Transaction) bool {
	if tx.To() != nil {
		_, ok := w.interestedAddrs.Load(tx.To().Hex())
		if ok {
			return true
		}
	}

	// check from
	from, err := w.getFromAddress(w.cfg.Chain, tx)
	utils.LogVerbose("from = ", from.Hex())
	if err == nil {
		_, ok := w.interestedAddrs.Load(from.Hex())
		if ok {
			return true
		}
	}

	return false
}

func (w *Watcher) getFromAddress(chain string, tx *etypes.Transaction) (common.Address, error) {
	signer := utils.GetEthChainSigner(chain)
	if signer == nil {
		return common.Address{}, fmt.Errorf("cannot find signer for chain %s", chain)
	}

	msg, err := tx.AsMessage(etypes.NewEIP2930Signer(tx.ChainId()))
	if err != nil {
		return common.Address{}, err
	}

	return msg.From(), nil
}

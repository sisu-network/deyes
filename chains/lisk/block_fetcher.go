package lisk

import (
	"fmt"
	"time"

	"github.com/sisu-network/deyes/utils"
	"go.uber.org/atomic"

	"github.com/sisu-network/deyes/chains/lisk/types"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/lib/log"
)

type BlockFetcher interface {
	start()
	stop()

	setBlockHeight()
	scanBlocks()
	getLatestBlock() (*types.Block, error)
	getBlock(height uint64) (*types.Block, error)
	tryGetBlock() (*types.Block, error)
	getBlockNumber() (uint64, error)
}

type BlockHeightExceededError struct {
	ChainHeight uint64
}

func NewBlockHeightExceededError(chainHeight uint64) error {
	return &BlockHeightExceededError{
		ChainHeight: chainHeight,
	}
}

func (e *BlockHeightExceededError) Error() string {
	return fmt.Sprintf("Our block height is higher than chain's height. Chain height = %d", e.ChainHeight)
}

type defaultBlockFetcher struct {
	blockHeight uint64
	blockTime   int
	cfg         config.Chain
	client      Client
	blockCh     chan *types.Block
	done        atomic.Bool
}

func newBlockFetcher(cfg config.Chain, blockCh chan *types.Block, client Client) BlockFetcher {
	return &defaultBlockFetcher{
		blockCh:   blockCh,
		cfg:       cfg,
		client:    client,
		blockTime: cfg.BlockTime,
		done:      *atomic.NewBool(false),
	}
}

func (bf *defaultBlockFetcher) start() {
	bf.setBlockHeight()
	bf.scanBlocks()
}

func (bf *defaultBlockFetcher) stop() {
	bf.done.Store(true)
}

func (bf *defaultBlockFetcher) setBlockHeight() {
	for {
		number, err := bf.getBlockNumber()
		if err != nil {
			log.Errorf("cannot get latest block number for chain %s. Sleeping for a few seconds", bf.cfg.Chain)
			time.Sleep(time.Second * 5)
			continue
		}

		bf.blockHeight = uint64(number)
		break
	}

	log.Info("Watching from block ", bf.blockHeight, " for chain ", bf.cfg.Chain)
}

func (bf *defaultBlockFetcher) scanBlocks() {
	latestBlock, err := bf.getLatestBlock()
	if err != nil {
		log.Error("Failed to scan blocks, err = ", err)
	}
	if latestBlock != nil {
		bf.blockHeight = latestBlock.Height
	}
	log.Info(bf.cfg.Chain, " Latest height = ", bf.blockHeight)

	for {
		log.Verbose("Block time on chain ", bf.cfg.Chain, " is ", bf.blockTime)

		// Get the block height
		block, err := bf.tryGetBlock()
		if err != nil || block == nil {
			bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
			time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
			continue
		}

		bf.blockTime = bf.blockTime - bf.cfg.AdjustTime/4
		bf.blockCh <- block
		bf.blockHeight++

		if bf.done.Load() {
			return
		}

		time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
	}
}

func (bf *defaultBlockFetcher) getLatestBlock() (*types.Block, error) {
	latestHeight, err := bf.client.BlockNumber()
	if err != nil {
		log.Errorf("Failed to get number by block, err = %v", err)
		return nil, err
	}

	return bf.getBlock(latestHeight)
}

func (bf *defaultBlockFetcher) getBlock(height uint64) (*types.Block, error) {
	block, err := bf.client.BlockByHeight(height)
	if err != nil {
		return nil, err
	}

	block.Transactions = []*types.Transaction{}
	if block.NumberOfTransactions > 0 {
		transactions, err := bf.client.TransactionByBlock(block.Id)
		if err != nil {
			log.Errorf("Failed to get transaction by block, err = %v", err)
			return nil, err
		}
		block.Transactions = transactions
	}

	return block, err
}

// Get block with retry when block is not mined yet.
func (bf *defaultBlockFetcher) tryGetBlock() (*types.Block, error) {
	number, err := bf.getBlockNumber()
	if err != nil {
		return nil, err
	}
	if number < uint64(bf.blockHeight) {
		return nil, NewBlockHeightExceededError(number)
	}

	block, err := bf.getBlock(bf.blockHeight)
	if err == nil {
		log.Verbose(bf.cfg.Chain, " Height = ", block.Height)
		return block, nil
	} else if err.Error() == "lisk block is not found" {
		// Sleep a few seconds and to get the block again.
		time.Sleep(time.Duration(utils.MinInt(bf.blockTime/4, 3000)) * time.Millisecond)
		block, err = bf.getBlock(bf.blockHeight)

		// Extend the wait time a bit more
		bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
		log.Verbose("New block-time: ", bf.blockTime)
	}

	return block, err
}

func (bf *defaultBlockFetcher) getBlockNumber() (uint64, error) {
	return bf.client.BlockNumber()
}

func (bf *defaultBlockFetcher) getAccount(address string) (*types.Account, error) {
	return bf.client.GetAccount(address)
}

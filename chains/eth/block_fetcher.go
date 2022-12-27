package eth

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
)

type defaultBlockFetcher struct {
	blockHeight int64
	blockTime   int
	cfg         config.Chain
	client      EthClient
	blockCh     chan *etypes.Block
}

func newBlockFetcher(cfg config.Chain, blockCh chan *etypes.Block, client EthClient) *defaultBlockFetcher {
	return &defaultBlockFetcher{
		blockCh:   blockCh,
		cfg:       cfg,
		client:    client,
		blockTime: cfg.BlockTime,
	}
}

func (bf *defaultBlockFetcher) start() {
	bf.setBlockHeight()

	bf.scanBlocks()
}

func (bf *defaultBlockFetcher) setBlockHeight() {
	for {
		number, err := bf.getBlockNumber()
		if err != nil {
			log.Errorf("cannot get latest block number for chain %s. Sleeping for a few seconds", bf.cfg.Chain)
			time.Sleep(time.Second * 5)
			continue
		}

		bf.blockHeight = int64(number)
		break
	}

	log.Info("Watching from block", bf.blockHeight, " for chain ", bf.cfg.Chain)
}

func (bf *defaultBlockFetcher) scanBlocks() {
	latestBlock, err := bf.getLatestBlock()
	if err != nil {
		log.Error("Failed to scan blocks, err = ", err)
	}

	if latestBlock != nil {
		bf.blockHeight = latestBlock.Header().Number.Int64()
	}
	log.Info(bf.cfg.Chain, " Latest height = ", bf.blockHeight)

	for {
		log.Verbose("Block time on chain ", bf.cfg.Chain, " is ", bf.blockTime)

		// Get the blockheight
		block, err := bf.tryGetBlock()
		if err != nil || block == nil {
			if _, ok := err.(*BlockHeightExceededError); !ok && err != ethereum.NotFound {
				// This err is not ETH not found or our custom error.
				log.Error("Cannot get block at height", bf.blockHeight, "for chain", bf.cfg.Chain, " err = ", err)

				// Bug only on polygon network https://github.com/maticnetwork/bor/issues/387
				// The block exists but its header hash is equivalent to empty root hash but the internal
				// block has some transaction inside. Geth client throws an error in this situation.
				// This rarely happens but it does happen. Skip this block for now.
				if strings.Index(bf.cfg.Chain, "polygon") >= 0 &&
					strings.Index(err.Error(), "server returned non-empty transaction list but block header indicates no transactions") >= 0 {
					bf.blockHeight = bf.blockHeight + 1
				}
			}

			bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
			time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
			continue
		}

		bf.blockTime = bf.blockTime - bf.cfg.AdjustTime/4
		bf.blockCh <- block
		bf.blockHeight++

		time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
	}
}

func (bf *defaultBlockFetcher) getLatestBlock() (*etypes.Block, error) {
	return bf.getBlock(-1)
}

func (bf *defaultBlockFetcher) getBlock(height int64) (*etypes.Block, error) {
	blockNum := big.NewInt(height)
	if height == -1 { // latest block
		blockNum = nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(bf.blockTime)*2*time.Millisecond)
	defer cancel()

	return bf.client.BlockByNumber(ctx, blockNum)
}

// Get block with retry when block is not mined yet.
func (bf *defaultBlockFetcher) tryGetBlock() (*etypes.Block, error) {
	number, err := bf.getBlockNumber()
	if err != nil {
		return nil, err
	}

	if number < uint64(bf.blockHeight) {
		return nil, NewBlockHeightExceededError(number)
	}

	block, err := bf.getBlock(bf.blockHeight)
	switch err {
	case nil:
		log.Verbose(bf.cfg.Chain, " Height = ", block.Number())
		return block, nil

	case ethereum.NotFound:
		// Sleep a few seconds and to get the block again.
		time.Sleep(time.Duration(utils.MinInt(bf.blockTime/4, 3000)) * time.Millisecond)
		block, err = bf.getBlock(bf.blockHeight)

		// Extend the wait time a little bit more
		bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
		log.Verbose("New blocktime: ", bf.blockTime)
	}

	return block, err
}

func (bf *defaultBlockFetcher) getBlockNumber() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(bf.blockTime)*2*time.Millisecond)
	defer cancel()

	return bf.client.BlockNumber(ctx)
}

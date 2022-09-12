package core

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
)

func TestBlockFetcher(t *testing.T) {
	t.Run("RPC should be successful if one RPC call fails and the other successful", func(t *testing.T) {
		expectedErr := fmt.Errorf("Cannot connect to RPC")
		expectedBlockNumber := uint64(10)
		expectedBlock := &etypes.Block{}
		expectedReceipt := &etypes.Receipt{}

		// Client1 does not work.
		client1 := &MockEthClient{
			BlockNumberFunc: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},

			BlockByNumberFunc: func(ctx context.Context, number *big.Int) (*etypes.Block, error) {
				return nil, expectedErr
			},

			TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
				return nil, expectedErr
			},
		}

		// Client2 works.
		client2 := &MockEthClient{
			BlockNumberFunc: func(ctx context.Context) (uint64, error) {
				return expectedBlockNumber, nil
			},

			BlockByNumberFunc: func(ctx context.Context, number *big.Int) (*etypes.Block, error) {
				return expectedBlock, nil
			},

			TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
				return expectedReceipt, nil
			},
		}

		fetcher := newBlockFetcher(config.Chain{
			Chain: "ganache1",
		}, nil, []EthClient{client1, client2})

		blockNumber, err := fetcher.getBlockNumber()
		require.Equal(t, nil, err)
		require.Equal(t, expectedBlockNumber, blockNumber)

		block, err := fetcher.getBlock(-1)
		require.Equal(t, nil, err)
		require.Equal(t, expectedBlock, block)
	})

	t.Run("RPC fails if all clients returns error", func(t *testing.T) {
		// Client1 does not work.
		client1 := &MockEthClient{
			BlockNumberFunc: func(ctx context.Context) (uint64, error) {
				return 0, fmt.Errorf("Cannot connect to RPC")
			},
		}

		// Client2 works.
		client2 := &MockEthClient{
			BlockNumberFunc: func(ctx context.Context) (uint64, error) {
				return 0, fmt.Errorf("Cannot connect to RPC")
			},
		}

		fetcher := newBlockFetcher(config.Chain{
			Chain: "ganache1",
		}, nil, []EthClient{client1, client2})

		blockNumber, err := fetcher.getBlockNumber()
		require.NotEqual(t, nil, err)
		require.Equal(t, uint64(0), blockNumber)
	})
}

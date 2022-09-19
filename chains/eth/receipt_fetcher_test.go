package eth

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestReceiptFetcher(t *testing.T) {
	t.Run("get_response_success", func(t *testing.T) {
		client := &MockEthClient{
			TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
				return &etypes.Receipt{}, nil
			},
		}

		fetcher := newReceiptFetcher(nil, []EthClient{client}, "ganache1").(*defaultReceiptFetcher)

		blockHeight := 12
		blockHash := "hash_12"
		response := fetcher.getResponse(&txReceiptRequest{
			blockNumber: int64(blockHeight),
			blockHash:   blockHash,
			txs: []*etypes.Transaction{
				etypes.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil),
			},
		})

		require.Equal(t, int64(blockHeight), response.blockNumber)
		require.Equal(t, blockHash, response.blockHash)
		require.Equal(t, 1, len(response.receipts))
		require.Equal(t, 1, len(response.txs))
	})

	t.Run("retry_succeeds", func(t *testing.T) {
		callCount := 0
		client := &MockEthClient{
			TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
				if callCount == 0 {
					callCount += 1
					return nil, fmt.Errorf("Cannot find receipt")
				}

				return &etypes.Receipt{}, nil
			},
		}

		fetcher := newReceiptFetcher(nil, []EthClient{client}, "ganache1").(*defaultReceiptFetcher)
		fetcher.retryTime = 0
		response := fetcher.getResponse(&txReceiptRequest{
			txs: []*etypes.Transaction{
				etypes.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil),
			},
		})

		require.Equal(t, 1, callCount)
		require.Equal(t, len(response.receipts), 1)
		require.Equal(t, len(response.txs), 1)
	})

	t.Run("retry_fails", func(t *testing.T) {
		callCount := 0
		client := &MockEthClient{
			TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
				callCount++
				return nil, fmt.Errorf("Cannot find receipt")
			},
		}

		fetcher := newReceiptFetcher(nil, []EthClient{client}, "ganache1").(*defaultReceiptFetcher)
		fetcher.retryTime = 0
		response := fetcher.getResponse(&txReceiptRequest{
			txs: []*etypes.Transaction{
				etypes.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil),
			},
		})

		require.Equal(t, MaxReceiptRetry+1, callCount)
		require.Equal(t, len(response.receipts), 0)
		require.Equal(t, len(response.txs), 0)
	})
}

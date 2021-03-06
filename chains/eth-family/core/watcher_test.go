package core

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	libchain "github.com/sisu-network/lib/chain"

	chainstypes "github.com/sisu-network/deyes/chains/types"

	"github.com/sisu-network/deyes/types"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/stretchr/testify/require"
)

func getTestDb() database.Database {
	db := database.NewDb(&config.Deyes{InMemory: true, DbHost: "localhost"})
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func TestProcessBlock(t *testing.T) {
	t.Parallel()

	client := &MockEthClient{
		PendingNonceAtFunc: func(ctx context.Context, account common.Address) (uint64, error) {
			return 0, nil
		},
		TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
			return &etypes.Receipt{
				Status: 1,
			}, nil
		},
	}

	db := getTestDb()
	cfg := config.Chain{
		Chain: "ganache1",
	}
	watcher := NewWatcher(db, cfg, make(chan *types.Txs), make(chan *chainstypes.TrackUpdate),
		[]EthClient{client}).(*Watcher)
	watcher.SetGateway(common.Address{1}.Hex())
	trans := genTransactions(t)
	hdr := etypes.Header{
		Difficulty: big.NewInt(100),
	}

	block := etypes.NewBlock(&hdr, trans, nil, nil, &mockTrieHasher{})
	txs, err := watcher.processBlock(block)
	require.NoError(t, err)
	require.NotNil(t, txs)
	require.Len(t, txs.Arr, len(trans))
}

func genTransactions(t *testing.T) etypes.Transactions {
	tx := etypes.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil)

	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	if err != nil {
		t.Fail()
		return nil
	}

	signedTx, err := etypes.SignTx(tx, libchain.GetEthChainSigner("ganache1"), privateKey)
	if err != nil {
		t.Fail()
		return nil
	}

	return etypes.Transactions{
		signedTx,
	}
}

func TestMultipleRpcs(t *testing.T) {
	t.Parallel()

	t.Run("RPC should be successful if one RPC call fails and the other successful", func(t *testing.T) {
		t.Parallel()

		expectedErr := fmt.Errorf("Cannot connect to RPC")
		expectedBlockNumber := uint64(10)
		expectedBlock := &etypes.Block{}
		expectedReceipt := &etypes.Receipt{}
		expectedGasPrice := big.NewInt(10)
		expectedNonce := uint64(10)

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

			SuggestGasPriceFunc: func(ctx context.Context) (*big.Int, error) {
				return nil, expectedErr
			},

			PendingNonceAtFunc: func(ctx context.Context, account common.Address) (uint64, error) {
				return 0, expectedErr
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

			SuggestGasPriceFunc: func(ctx context.Context) (*big.Int, error) {
				return expectedGasPrice, nil
			},

			PendingNonceAtFunc: func(ctx context.Context, account common.Address) (uint64, error) {
				return expectedNonce, nil
			},
		}

		watcher := Watcher{
			clients: []EthClient{client1, client2},
			cfg: config.Chain{
				Chain: "ganache1",
			},
		}

		blockNumber, err := watcher.getBlockNumber()
		require.Equal(t, nil, err)
		require.Equal(t, expectedBlockNumber, blockNumber)

		block, err := watcher.getBlock(-1)
		require.Equal(t, nil, err)
		require.Equal(t, expectedBlock, block)

		receipt, err := watcher.getTxReceipt(common.Hash{})
		require.Equal(t, nil, err)
		require.Equal(t, expectedReceipt, receipt)

		gasPrice, err := watcher.getSuggestedGasPrice()
		require.Equal(t, nil, err)
		require.Equal(t, expectedGasPrice, gasPrice)

		nonce := uint64(watcher.GetNonce("0x123"))
		require.Equal(t, expectedNonce, nonce)
	})

	t.Run("RPC fails if all clients returns error", func(t *testing.T) {
		t.Parallel()

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

		watcher := Watcher{
			clients: []EthClient{client1, client2},
			cfg: config.Chain{
				Chain: "ganache1",
			},
		}

		blockNumber, err := watcher.getBlockNumber()
		require.NotEqual(t, nil, err)
		require.Equal(t, uint64(0), blockNumber)
	})
}

package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	libchain "github.com/sisu-network/lib/chain"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
)

func TestProcessBlock(t *testing.T) {
	t.Parallel()

	client := &MockEthClient{
		TransactionReceiptFunc: func(ctx context.Context, txHash common.Hash) (*etypes.Receipt, error) {
			return &etypes.Receipt{
				Status: 1,
			}, nil
		},
	}

	watcher := Watcher{
		interestedAddrs: &sync.Map{},
		clients:         []EthClient{client},
		cfg: config.Chain{
			Chain: "ganache1",
		},
	}
	watcher.interestedAddrs.Store(common.Address{1}.Hex(), true)
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
		interestedAddrs: &sync.Map{},
		clients:         []EthClient{client1, client2},
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
}

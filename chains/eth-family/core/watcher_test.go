package core

import (
	"context"
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

type mockTrieHasher struct{}

func (h *mockTrieHasher) Reset() {}

func (h *mockTrieHasher) Update([]byte, []byte) {}

func (h *mockTrieHasher) Hash() common.Hash {
	return [32]byte{}
}

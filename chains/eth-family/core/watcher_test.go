package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestProcessBlock(t *testing.T) {
	t.Parallel()

	watcher := Watcher{}
	trans := genTransactions()
	hdr := etypes.Header{
		Difficulty: big.NewInt(100),
	}

	block := etypes.NewBlock(&hdr, trans, nil, nil, &mockTrieHasher{})
	txs, err := watcher.processBlock(block)
	require.NoError(t, err)
	require.NotNil(t, txs)
	require.Len(t, txs.Arr, len(trans))
}

func genTransactions() etypes.Transactions {
	tx := etypes.NewTransaction(0, common.Address{1}, big.NewInt(1), 22000, big.NewInt(1), nil)
	return etypes.Transactions{
		tx,
	}
}

type mockTrieHasher struct{}

func (h *mockTrieHasher) Reset() {}

func (h *mockTrieHasher) Update([]byte, []byte) {}

func (h *mockTrieHasher) Hash() common.Hash {
	return [32]byte{}
}

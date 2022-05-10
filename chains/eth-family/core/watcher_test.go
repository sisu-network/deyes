package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestProcessBlock(t *testing.T) {
	t.Parallel()

	watcher := Watcher{
		interestedAddrs: &sync.Map{},
	}
	watcher.interestedAddrs.Store(common.Address{1}.Hex(), true)
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

func TestGetBlock(t *testing.T) {
	client, err := ethclient.Dial("https://polygon-mumbai.infura.io/v3/b5c7d3d86ce341069cfbfd0c8714aab3")
	if err != nil {
		panic(err)
	}

	block, err := client.BlockByNumber(context.Background(), big.NewInt(26261184))
	if err != nil {
		panic(err)
	}

	fmt.Println("Tx length = ", block.Transactions().Len())
}

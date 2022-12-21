package eth

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	libchain "github.com/sisu-network/lib/chain"
	"github.com/stretchr/testify/require"

	etypes "github.com/ethereum/go-ethereum/core/types"
)

func getTestDb() database.Database {
	db := database.NewDb(&config.Deyes{InMemory: true, DbHost: "localhost"})
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func TestWatcher_TestProcessBlock(t *testing.T) {
	client := &MockEthClient{
		PendingNonceAtFunc: func(ctx context.Context, account common.Address) (uint64, error) {
			return 0, nil
		},
	}

	db := getTestDb()
	cfg := config.Chain{
		Chain: "ganache1",
	}
	watcher := NewWatcher(db, cfg, make(chan *types.Txs), make(chan *chainstypes.TrackUpdate),
		client).(*Watcher)

	gateway := common.Address{1}
	watcher.SetVault(gateway.Hex(), "")

	trans := []*etypes.Transaction{
		signTx(t, etypes.NewTransaction(0, gateway, big.NewInt(1), 22000, big.NewInt(1), nil)),
	}

	hdr := etypes.Header{
		Difficulty: big.NewInt(100),
	}

	block := etypes.NewBlock(&hdr, trans, nil, nil, &mockTrieHasher{})
	txs := watcher.processBlock(block)
	require.Equal(t, txs, trans)
}

func signTx(t *testing.T, tx *etypes.Transaction) *etypes.Transaction {
	privateKey, err := crypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.Nil(t, err)

	signedTx, err := etypes.SignTx(tx, libchain.GetEthChainSigner("ganache1"), privateKey)
	require.Nil(t, err)

	return signedTx
}

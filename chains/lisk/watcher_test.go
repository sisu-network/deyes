package lisk

import (
	"testing"

	ltype "github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
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

func TestWatcher_TestScanBlocks(t *testing.T) {
	vaultAddress := "lskcjqwdg9ezqtnpyd3f866nu4pdspsrft3rx5y8d"
	client := &MockLiskClient{
		BlockNumberFunc: func() (uint64, error) {
			return 1, nil
		},
		BlockByHeightFunc: func(height uint64) (*ltype.Block, error) {
			block := ltype.Block{
				Id:                   "mock_block_id",
				Height:               1,
				NumberOfTransactions: 1,
			}
			return &block, nil
		},
		TransactionByBlockFunc: func(block string) ([]ltype.Transaction, error) {
			sender := ltype.Sender{
				Address: "mock_sender_address",
			}
			asset := ltype.Asset{
				Amount: "1000",
				Data:   "mock_transaction_message",
				Recipient: ltype.AssetRecipient{
					Address: "mock_recipient_address",
				},
			}
			transaction := ltype.Transaction{
				Id:     "mock_transaction_id",
				Height: 1,
				Sender: sender,
				Asset:  asset,
			}
			return []ltype.Transaction{transaction}, nil
		},
	}

	db := getTestDb()
	cfg := config.Chain{
		Chain:      "lisk-testnet",
		BlockTime:  5000,
		AdjustTime: 1000,
		Rpcs:       []string{"https://testnet-service.lisk.com/api/v2"},
	}
	txsCh := make(chan *types.Txs)
	watcher := NewWatcher(db, cfg, txsCh, client).(*Watcher)
	watcher.SetVault(vaultAddress, "")
	block, _ := watcher.blockFetcher.tryGetBlock()
	require.Equal(t, block.Height, uint64(1))
	require.Equal(t, block.NumberOfTransactions, int64(1))
	require.Equal(t, len(block.Transactions), 1)
	require.Equal(t, watcher.vault, vaultAddress)

}

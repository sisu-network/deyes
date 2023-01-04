package lisk

import (
	"encoding/json"
	"testing"

	ltypes "github.com/sisu-network/deyes/chains/lisk/types"
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
		BlockByHeightFunc: func(height uint64) (*ltypes.Block, error) {
			block := ltypes.Block{
				Id:                   "mock_block_id",
				Height:               1,
				NumberOfTransactions: 1,
			}

			return &block, nil
		},
		TransactionByBlockFunc: func(block string) ([]*ltypes.Transaction, error) {
			sender := &ltypes.Sender{
				Address: "sender",
			}
			asset := &ltypes.Asset{
				Amount: "1000",
				Data:   "mock_transaction_message",
				Recipient: &ltypes.AssetRecipient{
					Address: vaultAddress,
				},
			}
			transaction := &ltypes.Transaction{
				Id:         "mock_transaction_id",
				Height:     1,
				Sender:     sender,
				Asset:      asset,
				Signatures: []string{"signature"},
			}

			return []*ltypes.Transaction{transaction}, nil
		},
	}

	db := getTestDb()
	cfg := config.Chain{
		Chain:      "lisk-testnet",
		BlockTime:  1000,
		AdjustTime: 100,
		Rpcs:       []string{"https://example.com"},
	}
	txsCh := make(chan *types.Txs)

	watcher := NewWatcher(db, cfg, txsCh, client).(*Watcher)
	watcher.SetVault(vaultAddress, "")
	watcher.Start()

	txs := <-txsCh
	require.Equal(t, 1, len(txs.Arr))

	tx := ltypes.Transaction{}
	err := json.Unmarshal(txs.Arr[0].Serialized, &tx)
	require.Nil(t, err)
	// TODO: Reconstruct the transaction and do verification for all fields.

	// Stop the watcher to clean up all running go routine.
	watcher.Stop()
}

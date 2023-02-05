package core

import (
	"sync"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/core/oracle"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

func mockForProcessor() (config.Deyes, database.Database, *MockClient, oracle.TokenPriceManager) {
	cfg := config.Deyes{
		DbHost:   "127.0.0.1",
		DbSchema: "deyes",
		InMemory: true,

		Chains: map[string]config.Chain{
			"ganache1": {
				Chain:     "ganache1",
				BlockTime: 1,
				Rpcs:      []string{"http://localhost:7545"},
			},
			"ganache2": {
				Chain:     "ganache2",
				BlockTime: 1,
				Rpcs:      []string{"http://localhost:8545"},
			},
		},
	}

	db := database.NewDb(&cfg)
	err := db.Init()
	if err != nil {
		panic(err)
	}

	networkHttp := network.NewHttp()
	sisuClient := &MockClient{}

	priceManager := oracle.NewTokenPriceManager(cfg, networkHttp)

	return cfg, db, sisuClient, priceManager

}

func TestProcessor(t *testing.T) {
	// TODO: Fix the in-memory db migration to bring back this test.
	t.Skip()

	t.Run("add_watcher_and_dispatcher", func(t *testing.T) {
		cfg, db, sisuClient, priceManager := mockForProcessor()
		processor := NewProcessor(&cfg, db, sisuClient, priceManager)
		processor.SetSisuReady(true)
		processor.Start()

		require.Equal(t, 2, len(processor.watchers))
		require.Equal(t, 2, len(processor.dispatchers))
	})

	t.Run("listen_txs_channel", func(t *testing.T) {
		cfg, db, sisuClient, priceManager := mockForProcessor()
		done := &sync.WaitGroup{}
		done.Add(1)

		sisuClient.BroadcastTxsFunc = func(txs *types.Txs) error {
			require.NotNil(t, txs)
			done.Done()
			return nil
		}

		processor := NewProcessor(&cfg, db, sisuClient, priceManager)
		processor.SetSisuReady(true)
		processor.Start()

		txs := &types.Txs{
			Chain: "ganache1",
			Block: 1,
			Arr:   make([]*types.Tx, 0),
		}

		processor.txsCh <- txs
		done.Wait()
	})
}

package main

import (
	"github.com/sisu-network/deyes/chains"
	chainlisk "github.com/sisu-network/deyes/chains/lisk"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
)

func testWatcher() {

	chainCfg := config.Chain{
		Chain:      "lisk-testnet",
		BlockTime:  20 * 1000,
		AdjustTime: 2000,
		Rpcs:       []string{"https://testnet-service.lisk.com/api/v2"},
	}

	cfg := config.Deyes{
		DbHost:   "127.0.0.1",
		DbSchema: "deyes",
		InMemory: true,

		Chains: map[string]config.Chain{"cardano-testnet": chainCfg},
	}

	dbInstance := database.NewDb(&cfg)
	err := dbInstance.Init()
	if err != nil {
		panic(err)
	}
	var watcher chains.Watcher
	client := chainlisk.NewLiskClient(chainCfg)
	watcher = chainlisk.NewWatcher(dbInstance, chainCfg, client)
	watcher.SetVault("lsk7kb3tbtrq5t4vbuzbavdahgj7yojmn4qbof63q", "")
	watcher.Start()

}

func main() {
	testWatcher()
}

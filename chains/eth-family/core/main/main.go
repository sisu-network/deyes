package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sisu-network/deyes/chains/eth-family/core"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
)

func getTestDb() database.Database {
	db := database.NewDb(&config.Deyes{InMemory: true, DbHost: "localhost"})
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func main() {
	db := getTestDb()
	rpcs := []string{
		"https://rpc.ankr.com/fantom_testnet/",
		// "https://xapi.testnet.fantom.network/lachesis/",
		// "https://polygon-testnet.blastapi.io/2f96b082-a1a4-4a39-839d-d95c8541e95f",
		// "https://rpc.testnet.fantom.network",
		// "https://fantom-testnet.blastapi.io/2f96b082-a1a4-4a39-839d-d95c8541e95f",
	}
	clients := core.NewEthClients(rpcs)
	// block, err := clients[0].BlockByNumber(context.Background(), big.NewInt(9829061))
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("Block = ", block.Header().Number)

	txsCh := make(chan *types.Txs, 1000)
	txTrackCh := make(chan *chainstypes.TrackUpdate, 1000)

	chainCfg := config.Chain{
		Chain:      "fantom-testnet",
		BlockTime:  5000,
		AdjustTime: 1000,
	}
	watcher := core.NewWatcher(db, chainCfg, txsCh, txTrackCh, clients)
	watcher.Start()

	go func() {
		for {
			select {
			case txs := <-txsCh:
				fmt.Println("There is a txs ", txs)
			}
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

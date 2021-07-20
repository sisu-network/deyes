package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	ethCore "github.com/sisu-network/deyes/chains/eth-family/core"
	"github.com/sisu-network/deyes/database"
)

func initializeDb() *database.Database {
	// Connect DB
	db := database.NewDb()
	err := db.Connect()
	if err != nil {
		panic(err)
	}

	// Migration
	err = db.DoMigration()
	if err != nil {
		panic(err)
	}

	return db
}

func initialize() {
	db := initializeDb()

	blockTimeString := os.Getenv("BLOCK_TIME")
	blockTime, err := strconv.Atoi(blockTimeString)
	if err != nil {
		panic(err)
	}

	chain := os.Getenv("CHAIN")
	fmt.Println("chain = ", chain)

	switch chain {
	case "eth":
		client := ethCore.NewClient(db, os.Getenv("CHAIN_RPC_URL"), blockTime, chain)
		client.Start()

	default:
		panic(fmt.Errorf("Unknown chain"))
	}

}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	initialize()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

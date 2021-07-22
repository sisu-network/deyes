package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/database"
)

func initializeDb() *database.Database {
	db := database.NewDb()
	err := db.Init()
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

	sisuUrl := os.Getenv("SISU_SERVER_URL")
	sisuClient := client.NewClient(sisuUrl)

	txProcessor := chains.NewTxProcessor(chain, blockTime, db, sisuClient)
	txProcessor.Start()
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

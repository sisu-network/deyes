package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/joho/godotenv"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/server"
	"github.com/sisu-network/deyes/utils"
)

func initializeDb() database.Database {
	db := database.NewDb()
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func setupApiServer(txProcessor *chains.TxProcessor) {
	handler := rpc.NewServer()
	handler.RegisterName("deyes", server.NewApi(txProcessor))

	if port, err := strconv.Atoi(os.Getenv("SERVER_PORT")); err != nil {
		panic(err)
	} else {
		s := server.NewServer(handler, port)
		s.Run()
	}
}

func initialize() {
	db := initializeDb()

	blockTimeString := os.Getenv("BLOCK_TIME")
	blockTime, err := strconv.Atoi(blockTimeString)
	if err != nil {
		panic(err)
	}

	chain := os.Getenv("CHAIN")
	utils.LogInfo("chain from config = ", chain)

	sisuUrl := os.Getenv("SISU_SERVER_URL")
	sisuClient := client.NewClient(sisuUrl)
	go sisuClient.TryDial()

	txProcessor := chains.NewTxProcessor(chain, blockTime, db, sisuClient)
	txProcessor.Start()

	setupApiServer(txProcessor)
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

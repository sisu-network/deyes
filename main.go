package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/server"
)

func initializeDb(cfg *config.Deyes) database.Database {
	db := database.NewDb(cfg)
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func setupApiServer(cfg *config.Deyes, txProcessor *chains.TxProcessor) {
	handler := rpc.NewServer()
	handler.RegisterName("deyes", server.NewApi(txProcessor))

	s := server.NewServer(handler, cfg.ServerPort)
	s.Run()
}

func initialize(cfg *config.Deyes) {
	db := initializeDb(cfg)

	sisuClient := client.NewClient(cfg.SisuServerUrl)
	go sisuClient.TryDial()

	txProcessor := chains.NewTxProcessor(cfg, db, sisuClient)
	txProcessor.Start()

	setupApiServer(cfg, txProcessor)
}

func writeDefaultConfig(filePath string) error {
	err := ioutil.WriteFile(filePath, []byte(config.EyesConfigTemplate), 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfig() *config.Deyes {
	tomlFile := "./deyes.toml"
	if _, err := os.Stat(tomlFile); os.IsNotExist(err) {
		panic(err)
	}

	cfg := new(config.Deyes)
	_, err := toml.DecodeFile(tomlFile, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func main() {
	cfg := loadConfig()

	initialize(cfg)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/sisu-network/deyes/core/oracle/sushiswap"
	"github.com/sisu-network/deyes/core/oracle/uniswap"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/logdna/logdna-go/logger"
	"github.com/sisu-network/deyes/client"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/core"
	"github.com/sisu-network/deyes/core/oracle"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/network"
	"github.com/sisu-network/deyes/server"
	"github.com/sisu-network/lib/log"
)

func initializeDb(cfg *config.Deyes) database.Database {
	db := database.NewDb(cfg)
	err := db.Init()
	if err != nil {
		panic(err)
	}

	return db
}

func setupApiServer(cfg *config.Deyes, processor *core.Processor) {
	handler := rpc.NewServer()
	handler.RegisterName("deyes", server.NewApi(processor))

	log.Info("Running server at port", cfg.ServerPort)
	s := server.NewServer(handler, cfg.ServerPort)
	s.Run()
}

func initialize(cfg *config.Deyes) {
	db := initializeDb(cfg)

	sisuClient := client.NewClient(cfg.SisuServerUrl)
	go sisuClient.TryDial()

	networkHttp := network.NewHttp()
	uniswapManager := uniswap.NewUniwapManager(*cfg)
	sushiSwapManager := sushiswap.NewSushiSwapManager(*cfg)
	priceManager := oracle.NewTokenPriceManager(*cfg, networkHttp, uniswapManager, sushiSwapManager)

	processor := core.NewProcessor(cfg, db, sisuClient, priceManager)
	processor.Start()

	setupApiServer(cfg, processor)
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
	if len(cfg.LogDNA.Secret) > 0 {
		opts := logger.Options{
			App:           cfg.LogDNA.AppName,
			FlushInterval: cfg.LogDNA.FlushInterval.Duration,
			Hostname:      cfg.LogDNA.HostName,
			MaxBufferLen:  cfg.LogDNA.MaxBufferLen,
		}
		logDNA := log.NewDNALogger(cfg.LogDNA.Secret, opts, false)
		log.SetLogger(logDNA)
	}

	initialize(cfg)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

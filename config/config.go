package config

import (
	"github.com/sisu-network/lib/log"
)

var ChainParamsMap = map[string]ChainParams{
	"ganache1":         {GasPriceStartBlockHeight: 1000, Interval: 50},
	"ganache2":         {GasPriceStartBlockHeight: 1000, Interval: 50},
	"eth":              {GasPriceStartBlockHeight: 1000, Interval: 50},
	"ropsten-testnet":  {GasPriceStartBlockHeight: 1000, Interval: 50},
	"goerli-testnet":   {GasPriceStartBlockHeight: 1000, Interval: 50},
	"binance-testnet":  {GasPriceStartBlockHeight: 1000, Interval: 50},
	"fantom-testnet":   {GasPriceStartBlockHeight: 1000, Interval: 50},
	"polygon-testnet":  {GasPriceStartBlockHeight: 1000, Interval: 50},
	"xdai":             {GasPriceStartBlockHeight: 1000, Interval: 50},
	"arbitrum-testnet": {GasPriceStartBlockHeight: 1000, Interval: 50},
}

type ChainParams struct {
	GasPriceStartBlockHeight int64
	Interval                 int64
}

type Chain struct {
	Chain      string   `toml:"chain" json:"chain"`
	BlockTime  int      `toml:"block_time" json:"block_time"`
	AdjustTime int      `toml:"adjust_time" json:"adjust_time"`
	Rpcs       []string `toml:"rpcs" json:"rpcs"`
	RpcSecret  string   `toml:"rpc_secret" json:"rpc_secret"`
}

type Deyes struct {
	DbHost     string `toml:"db_host"`
	DbPort     int    `toml:"db_port"`
	DbUsername string `toml:"db_username"`
	DbPassword string `toml:"db_password"`
	DbSchema   string `toml:"db_schema"`

	PriceOracleUrl     string   `toml:"price_oracle_url"`
	PriceOracleSecret  string   `toml:"price_oracle_secret"`
	PricePollFrequency int      `toml:"price_poll_frequency"`
	PriceTokenList     []string `toml:"price_token_list"`

	ServerPort    int    `toml:"server_port"`
	SisuServerUrl string `toml:"sisu_server_url"`

	Chains map[string]Chain `toml:"chains"`

	LogDNA log.LogDNAConfig `toml:"log_dna"`

	InMemory bool // Used in test only
}

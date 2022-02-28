package config

import "time"

var ChainParamsMap = map[string]ChainParams{
	"ganache1": {GasPriceStartBlockHeight: 1000, Interval: 50},
	"ganache2": {GasPriceStartBlockHeight: 1000, Interval: 50},
}

type ChainParams struct {
	GasPriceStartBlockHeight int64
	Interval                 int64
}

type Chain struct {
	Chain     string `toml:"chain"`
	BlockTime int    `toml:"block_time"`
	RpcUrl    string `toml:"rpc_url"`
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

	LogDNA LogDNA `toml:"log_dna"`
}

type LogDNA struct {
	Secret        string        `toml:"secret"`
	AppName       string        `toml:"app_name"`
	HostName      string        `toml:"host_name"`
	FlushInterval time.Duration `toml:"flush_interval"`
	MaxBufferLen  int           `toml:"max_buffer_len"`
	Level         string        `toml:"level"`
}

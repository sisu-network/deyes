package config

var ChainParamsMap = map[string]ChainParams{
	"ganache1":            {GasPriceStartBlockHeight: 1000, Interval: 10},
	"ganache2":            {GasPriceStartBlockHeight: 1000, Interval: 10},
	"eth-ropsten":         {GasPriceStartBlockHeight: 1000, Interval: 10},
	"eth-binance-testnet": {GasPriceStartBlockHeight: 1000, Interval: 10},
	"fantom-testnet":      {GasPriceStartBlockHeight: 1000, Interval: 10},
	"polygon-testnet":     {GasPriceStartBlockHeight: 1000, Interval: 10},
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

	ServerPort    int    `toml:"server_port"`
	SisuServerUrl string `toml:"sisu_server_url"`

	Chains map[string]Chain `toml:"chains"`
}

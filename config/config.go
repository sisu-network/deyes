package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/sisu-network/lib/log"
)

type SyncDbConfig struct {
	Host      string `toml:"host" json:"host,omitempty"`
	Port      int    `toml:"port" json:"port,omitempty"`
	User      string `toml:"user" json:"user,omitempty"`
	Password  string `toml:"password" json:"password,omitempty"`
	DbName    string `toml:"db_name" json:"db_name,omitempty"`
	SubmitURL string `toml:"submit_url" json:"submit_url,omitempty"`
}

type ClientType string

const (
	ClientTypeBlockFrost ClientType = "block_frost"
	ClientTypeSelfHost   ClientType = "self_host"
)

type Chain struct {
	Chain      string   `toml:"chain" json:"chain"`
	BlockTime  int      `toml:"block_time" json:"block_time"`
	AdjustTime int      `toml:"adjust_time" json:"adjust_time"`
	Rpcs       []string `toml:"rpcs" json:"rpcs"`
	Wss        []string `toml:"wss" json:"wss"`

	// ETH
	UseEip1559 bool `toml:"use_eip_1559" json:"use_eip_1559"` // For gas calculation

	// Cardano
	ClientType ClientType `toml:"client_type" json:"client_type"`
	RpcSecret  string     `toml:"rpc_secret" json:"rpc_secret"`
	// SyncDB config
	SyncDB SyncDbConfig `toml:"sync_db" json:"sync_db"`

	// Solana
	SolanaBridgeProgramId string `toml:"solana_bridge_program_id" json:"solana_bridge_program_id"`
}

type Token struct {
	Symbol        string `toml:"symbol"`
	NameLowerCase string `toml:"name_lower_case"`
	Address       string `toml:"address"`
	ChainId       string `toml:"chain_id"`
	ChainName     string `toml:"chain_name"`
}

type PriceProvider struct {
	Url     string `toml:"url"`
	Secrets string `toml:"secrets"`
}

type Deyes struct {
	DbHost     string `toml:"db_host"`
	DbPort     int    `toml:"db_port"`
	DbUsername string `toml:"db_username"`
	DbPassword string `toml:"db_password"`
	DbSchema   string `toml:"db_schema"`

	PriceProviders map[string]PriceProvider `toml:"price_providers"`
	Tokens         map[string]Token         `toml:"tokens"`

	ServerPort    int    `toml:"server_port"`
	SisuServerUrl string `toml:"sisu_server_url"`

	// This variable indicates if we should use some external rpcs or not (in chainlist.org). If you
	// are running local node or you are certain that your node is always online, you don't need to
	// enable this variable. If your rpcs are unstable, you might want to turn on this variable at
	// limited time.
	UseExternalRpcsInfo bool `toml:"use_external_rpcs_info"`
	// Chains config
	Chains map[string]Chain `toml:"chains"`

	// LogDNA
	LogDNA log.LogDNAConfig `toml:"log_dna"`

	InMemory bool // Used in test only
}

func Load(path string) Deyes {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(err)
	}

	cfg := Deyes{}
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

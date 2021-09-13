package config

type Chain struct {
	Chain         string `toml:"chain"`
	BlockTime     int    `toml:"block_time"`
	StartingBlock int    `toml:"starting_block"`
	RpcUrl        string `toml:"rpc_url"`
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

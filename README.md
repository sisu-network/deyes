Deyes (Dragon eyes) is a repo that contains all chain watchers. Each watcher connects to a (RPC) end point to get transaction data from a blockchain.

The data source could be a local running node or a trusted party service. We leave this option to the node operator to choose the best choice for them.

## Run Deyes locally

Creates a `deyes.toml` in the root folder of this repo with the following content. Make sure that you have mysql installed with the following configs.

```
db_host = "localhost"
db_port = 3306
db_username = "root"
db_password = "password"
db_schema = "deyes"

server_port = 31001
sisu_server_url = "http://localhost:25456"

[chains]
[chains.eth]
  chain = "eth"
  block_time = 1000
  starting_block = 0
  rpc_url = "http://localhost:7545"

[chains.sisu-eth]
  chain = "sisu-eth"
  block_time = 1000
  starting_block = 0
  rpc_url = "http://localhost:8545"
```

Build and run deyes

```
go build && ./deyes
```

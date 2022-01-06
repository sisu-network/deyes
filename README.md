Deyes (Dragon eyes) is a repo that contains all chain watchers. Each watcher connects to a (RPC) end point to get transaction data from a blockchain.

The data source could be a local running node or a trusted party service. We leave this option to the node operator to choose the best choice for them.

## Run Deyes locally

Creates a `deyes.toml` in the root folder of this repo with the following content. Make sure that you have mysql installed with the following configs.

`cp deyes.toml.dev deyes.toml`

Install all modules.
```
go mod tidy
```

Build and run deyes

```
go build && ./deyes
```

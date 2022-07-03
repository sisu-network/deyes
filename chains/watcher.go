package chains

type Watcher interface {
	Start()

	// Set an account on a chain that Sisu controls. On some chains like ETH, the account is different
	// from the gateway contract.
	SetChainAccount(addr string)

	// Set gateway of the network. On chains like BTC, Cardano the gateway is the same as chain
	// account.
	SetGateway(addr string)
}

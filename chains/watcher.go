package chains

type Watcher interface {
	Start()

	// Set gateway of the network. On chains like BTC, Cardano the gateway is the same as chain
	// account.
	SetGateway(addr string)

	// Track a particular tx whose binary form on that chain is bz
	TrackTx(txHash string)
}

package chains

type Watcher interface {
	Start()

	// Set vault of the network. On chains like BTC, Cardano the gateway is the same as chain
	// account.
	SetVault(addr string)

	// Track a particular tx whose binary form on that chain is bz
	TrackTx(txHash string)
}

package chains

type Watcher interface {
	Start()

	// AddWatchAddr(addr string)
	SetGateway(addr string)
}

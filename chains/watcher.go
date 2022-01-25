package chains

type Watcher interface {
	Start()
	AddWatchAddr(addr string)
	GetNonce(address string) int64
	GetGasPrice() int64
}

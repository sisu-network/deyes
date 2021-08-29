package chains

type WatcherInterface interface {
	Start()
	AddWatchAddr(addr string)
}

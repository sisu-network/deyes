package chains

type Watcher interface {
	Start()
	AddWatchAddr(addr string)
}

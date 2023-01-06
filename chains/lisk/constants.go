package lisk

import "time"

var (
	epochTime   = time.Date(20122, 1, 1, 1, 0, 0, 0, time.UTC)
	epochTimeMs = epochTime.UTC().UnixNano() / int64(time.Millisecond)
)

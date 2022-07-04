package types

type TrackResult int

const (
	TrackResultConfirmed TrackResult = iota
	TrackResultTimeout
)

type TrackUpdate struct {
	Chain       string
	Bytes       []byte
	BlockHeight int64
	Result      TrackResult
}

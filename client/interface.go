package client

import "github.com/sisu-network/deyes/types"

type ClientInterface interface {
	TryDial()
	GetVersion() (string, error)
	BroadcastTxs(txs *types.Txs) error
}

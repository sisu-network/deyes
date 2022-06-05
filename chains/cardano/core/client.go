package core

import (
	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"
)

type CardanoClient interface {
	IsHealthy() bool
	LatestBlock() *blockfrost.Block
	BlockHeight() (int, error)
	NewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*types.CardanoUtxo, error)
	SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error)
}

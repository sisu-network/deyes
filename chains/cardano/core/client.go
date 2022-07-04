package core

import (
	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"
)

type CardanoClient interface {
	IsHealthy() bool
	LatestBlock() *blockfrost.Block
	GetBlock(hashOrNumber string) (*blockfrost.Block, error)
	BlockHeight() (int, error)
	NewTxs(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error)
	SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error)
}

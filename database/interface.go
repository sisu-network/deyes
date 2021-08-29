package database

import "github.com/sisu-network/deyes/types"

type DatabaseInterface interface {
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)
	LoadBlockHeight(chain string) (int64, error)
}

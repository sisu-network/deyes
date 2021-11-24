package database

import "github.com/sisu-network/deyes/types"

type MockDb struct {
	InitFunc               func() error
	SaveTxsFunc            func(chain string, blockHeight int64, txs *types.Txs)
	SaveWatchAddressFunc   func(chain, address string)
	LoadWatchAddressesFunc func(chain string) []string
}

func (mock *MockDb) Init() error {
	if mock.InitFunc != nil {
		return mock.InitFunc()
	}

	return nil
}

func (mock *MockDb) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	if mock.SaveTxsFunc != nil {
		mock.SaveTxsFunc(chain, blockHeight, txs)
	}
}

func (mock *MockDb) SaveWatchAddress(chain, address string) {
	if mock.SaveWatchAddressFunc != nil {
		mock.SaveWatchAddressFunc(chain, address)
	}
}
func (mock *MockDb) LoadWatchAddresses(chain string) []string {
	if mock.LoadWatchAddressesFunc != nil {
		return mock.LoadWatchAddressesFunc(chain)
	}

	return nil
}

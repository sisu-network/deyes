package chains

import (
	"context"

	eTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/utils"
)

type Dispatcher interface {
	Start()
	Dispatch(tx []byte) error
}

type DefaultDispatcher struct {
	chain, rpcEndpoint string
	client             *ethclient.Client
}

func NewDispatcher(chain, rpcEndpoint string) Dispatcher {
	return &DefaultDispatcher{
		chain:       chain,
		rpcEndpoint: rpcEndpoint,
	}
}

func (d *DefaultDispatcher) Start() {
	var err error
	d.client, err = ethclient.Dial(d.rpcEndpoint)
	if err != nil {
		utils.LogError("Cannot dial chain", d.chain, "at endpoint", d.rpcEndpoint)
		// TODO: Add retry mechanism here.
		return
	}
}

func (d *DefaultDispatcher) Dispatch(txBytes []byte) error {
	tx := &eTypes.Transaction{}
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return err
	}

	return d.client.SendTransaction(context.Background(), tx)
}

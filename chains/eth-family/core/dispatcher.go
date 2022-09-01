package core

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	eTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
)

type EthDispatcher struct {
	chain   string
	rpcs    []string
	clients []*ethclient.Client
	healthy []bool
}

func NewEhtDispatcher(chain string, rpcs []string) chains.Dispatcher {
	return &EthDispatcher{
		chain:   chain,
		rpcs:    rpcs,
		healthy: make([]bool, len(rpcs)),
		clients: make([]*ethclient.Client, len(rpcs)),
	}
}

func (d *EthDispatcher) Start() {
	d.dial()
}

func (d *EthDispatcher) dial() {
	var err error
	for i := range d.rpcs {
		if d.clients[i] != nil {
			d.clients[i].Close()
		}
		d.clients[i], err = ethclient.Dial(d.rpcs[i])
		if err != nil {
			log.Error("Cannot dial chain", d.chain, "at endpoint", d.rpcs[i])
			d.healthy[i] = false
			log.Infof("RPC %s is NOT healthy", d.rpcs[i])
		} else {
			d.healthy[i] = true
			log.Infof("RPC %s is healthy", d.rpcs[i])
		}
	}
}

func (d *EthDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	txBytes := request.Tx

	tx := &eTypes.Transaction{}
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		return types.NewDispatchTxError(err)
	}

	// Check if this is a contract deployment for eth. If it is, returns the deployed address.
	var addr string
	from := utils.PublicKeyBytesToAddress(request.PubKey)
	err = d.tryDispatchTx(tx, request.Chain, from)
	if err != nil {
		// Try to connect again.
		d.dial()
		err = d.tryDispatchTx(tx, request.Chain, from) // try second time
	}

	if err == nil {
		log.Verbose("Tx is dispatched successfully for chain ", request.Chain, " from ", from,
			"txHash =", tx.Hash())
		return &types.DispatchedTxResult{
			Success:      true,
			DeployedAddr: addr,
			Chain:        request.Chain,
			TxHash:       request.TxHash,
		}
	}

	return types.NewDispatchTxError(err)
}

func (d *EthDispatcher) tryDispatchTx(tx *eTypes.Transaction, chain string, from common.Address) error {
	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthy, rpcs := d.shuffle()

	for i := range clients {
		if !healthy[i] {
			log.Verbose("%s is not healthy", rpcs[i])
			continue
		}

		log.Verbose("Trying rpc ", rpcs[i])
		client := clients[i]
		if err := client.SendTransaction(context.Background(), tx); err != nil {
			// It is possible that another node has deployed the same transaction. We check if the tx has
			// been included into the blockchain or not.
			_, _, err2 := client.TransactionByHash(context.Background(), tx.Hash())
			if err2 != nil {
				log.Error("cannot dispatch tx, from = ", from, " chain = ", chain)
				log.Error("cannot dispatch tx, err = ", err)
				log.Error("cannot dispatch tx, err2 = ", err2)

				d.healthy[i] = false
				continue
			}

			log.Info("The transaction has been deployed before. Tx hash = ", tx.Hash().String())
			return nil
		}
	}

	return fmt.Errorf("cannot dispatch eth tx")
}

func (d *EthDispatcher) shuffle() ([]*ethclient.Client, []bool, []string) {
	n := len(d.clients)

	clients := make([]*ethclient.Client, n)
	healthy := make([]bool, n)
	rpcs := make([]string, n)

	copy(clients, d.clients)
	copy(healthy, d.healthy)
	copy(rpcs, d.rpcs)

	for i := 0; i < 20; i++ {
		x := rand.Intn(n)
		y := rand.Intn(n)

		tmpClient := clients[x]
		clients[x] = clients[y]
		clients[y] = tmpClient

		tmpHealth := healthy[x]
		healthy[x] = healthy[y]
		healthy[y] = tmpHealth

		tmpRpc := rpcs[x]
		rpcs[x] = rpcs[y]
		rpcs[y] = tmpRpc
	}

	return clients, healthy, rpcs
}

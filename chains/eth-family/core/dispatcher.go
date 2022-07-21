package core

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	eTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/crypto"
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
		healthy: make([]bool, 0),
	}
}

func (d *EthDispatcher) Start() {
	d.dial()
}

func (d *EthDispatcher) dial() {
	var err error
	for i := range d.rpcs {
		d.clients[i], err = ethclient.Dial(d.rpcs[0])
		if err != nil {
			log.Error("Cannot dial chain", d.chain, "at endpoint", d.rpcs[0])
			d.healthy[i] = false
		} else {
			d.healthy[i] = true
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
	if request.IsEthContractDeployment {
		if request.PubKey == nil {
			return types.NewDispatchTxError(fmt.Errorf("invalid pubkey"))
		}

		addr = crypto.CreateAddress(from, tx.Nonce()).String()

		log.Info("Deploying address = ", addr, " for chain ", request.Chain)
	}

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
			Success:                 true,
			DeployedAddr:            addr,
			Chain:                   request.Chain,
			TxHash:                  request.TxHash,
			IsEthContractDeployment: request.IsEthContractDeployment,
		}
	}

	return types.NewDispatchTxError(err)
}

func (d *EthDispatcher) tryDispatchTx(tx *eTypes.Transaction, chain string, from common.Address) error {
	for i := range d.clients {
		if !d.healthy[i] {
			continue
		}

		client := d.clients[i]
		if err := client.SendTransaction(context.Background(), tx); err != nil {
			// It is possible that another node has deployed the same transaction. We check if the tx has
			// been included into the blockchain or not.
			_, _, err2 := client.TransactionByHash(context.Background(), tx.Hash())
			if err2 != nil {
				log.Error("cannot dispatch tx, from = ", from, " chain = ", chain)
				log.Error("cannot dispatch tx, err = ", err)
				log.Error("cannot dispatch tx, err2 = ", err2)
				return err
			}

			log.Info("The transaction has been deployed before. Tx hash = ", tx.Hash().String())
		}
	}

	return fmt.Errorf("cannot dispatch eth tx")
}

package eth

import (
	"context"
	"fmt"
	"math/big"
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
		log.Error("Failed to unmarshal ETH transaction, err = ", err)
		return types.NewDispatchTxError(request, types.ErrMarshal)
	}

	from := utils.PublicKeyBytesToAddress(request.PubKey)
	// Check the balance to see if we have enough native token.
	balance := d.checkNativeBalance(from)
	err = nil
	if balance == nil {
		err = fmt.Errorf("Cannot get balance for account %s", from)
	}

	minimum := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(tx.Gas())))
	minimum = minimum.Add(minimum, tx.Value())
	if minimum.Cmp(balance) > 0 {
		err = fmt.Errorf("Balance smaller than minimum required for this transaction, from = %s, balance = %s, minimum = %s, chain = %s",
			from.String(), balance.String(), minimum.String(), request.Chain)
	}

	if err != nil {
		log.Error(err)
		return &types.DispatchedTxResult{
			Success: false,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
			Err:     types.ErrNotEnoughBalance,
		}
	}

	// Dispath tx.
	err = d.tryDispatchTx(tx, request.Chain, from)
	if err == nil {
		log.Verbose("Tx is dispatched successfully for chain ", request.Chain, " from ", from,
			" txHash =", tx.Hash())
		return &types.DispatchedTxResult{
			Success: true,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	} else {
		log.Error("Failed to dispatch tx, err = ", err)
	}

	return types.NewDispatchTxError(request, types.ErrSubmitTx)
}

func (d *EthDispatcher) checkNativeBalance(from common.Address) *big.Int {
	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthy, rpcs := d.shuffle()

	for i := range clients {
		if !healthy[i] {
			log.Verbose("%s is not healthy", rpcs[i])
			continue
		}

		client := clients[i]
		balance, err := client.BalanceAt(context.Background(), from, nil)
		if err != nil {
			log.Error("Error getting balance, err = ", err)
			continue
		}

		if balance.Cmp(big.NewInt(0)) == 0 {
			// Balance is rarely 0. If this is the case, it's very likely that this RPC has a problem.
			// Continue with other RPC.
			log.Warn("balance is 0 for some rpc. Trying other rpc")
			continue
		}

		return balance
	}

	return nil
}

func (d *EthDispatcher) tryDispatchTx(tx *eTypes.Transaction, chain string, from common.Address) error {
	// Shuffle rpcs so that we will use different healthy rpc
	clients, healthy, rpcs := d.shuffle()

	triedRpcs := make(map[string]bool)

	for i := range clients {
		if !healthy[i] {
			log.Verbose("%s is not healthy", rpcs[i])
			continue
		}

		log.Verbose("Trying rpc ", rpcs[i])
		err := d.sendTransaction(clients[i], tx, i, from, chain, rpcs[i])
		if err == nil {
			return nil
		}

		triedRpcs[rpcs[i]] = true
	}

	// We cannot connect to healthy rpcs. Try connecting to rpc that was marked as unhealthy in the
	// past.
	for i, rpc := range rpcs {
		if triedRpcs[rpc] {
			continue
		}

		log.Verbose("Trying rpc ", rpcs[i])
		err := d.sendTransaction(clients[i], tx, i, from, chain, rpcs[i])
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("cannot dispatch eth tx")
}

func (d *EthDispatcher) sendTransaction(client *ethclient.Client, tx *eTypes.Transaction, index int,
	from common.Address, chain string, rpc string) error {
	if err := client.SendTransaction(context.Background(), tx); err != nil {
		// It is possible that another node has deployed the same transaction. We check if the tx has
		// been included into the blockchain or not.
		_, _, err2 := client.TransactionByHash(context.Background(), tx.Hash())
		if err2 != nil {
			log.Error("cannot dispatch tx, from = ", from, " chain = ", chain, " rpc = ", rpc)
			log.Error("cannot dispatch tx, err = ", err)
			log.Error("cannot dispatch tx, err2 = ", err2)

			d.healthy[index] = false

			return err
		}

		d.healthy[index] = true

		log.Info("The transaction has been deployed before. Tx hash = ", tx.Hash().String())
		return nil
	}

	d.healthy[index] = true

	return nil
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
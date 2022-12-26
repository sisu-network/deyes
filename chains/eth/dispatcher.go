package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
	eTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
	"github.com/sisu-network/lib/log"
)

type EthDispatcher struct {
	chain  string
	client EthClient
}

func NewEhtDispatcher(chain string, client EthClient) chains.Dispatcher {
	return &EthDispatcher{
		chain:  chain,
		client: client,
	}
}

// Start implements Dispatcher interface.
func (d *EthDispatcher) Start() {
	// Do nothing.
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
	balance, err := d.client.BalanceAt(context.Background(), from, nil)
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
	resultErr := types.ErrNil
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
		switch err {
		case ethcore.ErrNonceTooLow:
			resultErr = types.ErrNonceTooLow
		default:
			resultErr = types.ErrSubmitTx
		}
	}

	return types.NewDispatchTxError(request, resultErr)
}

func (d *EthDispatcher) tryDispatchTx(tx *eTypes.Transaction, chain string, from common.Address) error {
	return d.client.SendTransaction(context.Background(), tx)
}

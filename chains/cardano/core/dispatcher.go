package core

import (
	"encoding/base64"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/chains"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type CardanoDispatcher struct {
	client CardanoClient
}

func NewDispatcher(client CardanoClient) chains.Dispatcher {
	return &CardanoDispatcher{
		client: client,
	}
}

func (d *CardanoDispatcher) Start() {
}

func (d *CardanoDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	log.Debug("Dispatching cardano transaction ...")
	bz := base64.StdEncoding.EncodeToString(request.Tx)
	log.Debug("bz when dispatch = ", bz)

	tx := &cardano.Tx{}
	if err := tx.UnmarshalCBOR(request.Tx); err != nil {
		log.Error("error when unmarshalling tx: ", err)
		return &types.DispatchedTxResult{
			Success: false,
			Err:     err,
		}
	}

	for _, txInput := range tx.Body.Inputs {
		txInput.Amount = nil
	}

	log.Debug("tx fee = ", tx.Body.Fee)

	hash, err := d.client.SubmitTx(tx)
	if err != nil {
		log.Error("error when submitting tx: ", err)
		return &types.DispatchedTxResult{
			Success: false,
			Err:     err,
		}
	}

	return &types.DispatchedTxResult{
		Success: true,
		TxHash:  hash.String(),
	}
}

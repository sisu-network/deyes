package core

import (
	"encoding/json"

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
	tx := &cardano.Tx{}
	// We are using json to marshal tx at the moment because the cbor's marshalling of tx is not ready yet.
	err := json.Unmarshal(request.Tx, tx)
	if err != nil {
		log.Error("error when unmarshalling tx", err)
		return &types.DispatchedTxResult{
			Success: false,
			Err:     err,
		}
	}

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

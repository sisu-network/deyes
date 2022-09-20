package cardano

import (
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
	if err := tx.UnmarshalCBOR(request.Tx); err != nil {
		log.Error("error when unmarshalling tx: ", err)
		return types.NewDispatchTxError(request, types.ErrMarshal)
	}

	for _, txInput := range tx.Body.Inputs {
		txInput.Amount = nil
	}

	hash, err := d.client.SubmitTx(tx)
	if err != nil {
		log.Error("error when submitting tx: ", err)
		return types.NewDispatchTxError(request, types.ErrSubmitTx)
	}

	log.Verbose("Cardano tx hash = ", hash)

	return &types.DispatchedTxResult{
		Success: true,
		Chain:   request.Chain,
		TxHash:  hash.String(),
	}
}

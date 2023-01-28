package solana

import (
	"context"
	"encoding/base64"
	"os"

	bin "github.com/gagliardetto/binary"
	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/text"
	"github.com/ybbus/jsonrpc/v3"

	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type Dispatcher struct {
	clientUrls []string
	clientRpcs []jsonrpc.RPCClient
}

func NewDispatcher(clientUrls, wsUrls []string) *Dispatcher {
	clientRpcs := make([]jsonrpc.RPCClient, 0)

	for _, url := range clientUrls {
		clientRpcs = append(clientRpcs, jsonrpc.NewClient(url))
	}
	return &Dispatcher{
		clientUrls: clientUrls,
		clientRpcs: clientRpcs,
	}
}

func (d *Dispatcher) Start() {
}

func (d *Dispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	for i, client := range d.clientRpcs {
		log.Verbosef("Dispatching solana tx using url %s", d.clientUrls[i])
		signature, err := d.sendTransaction(request.Tx, client)
		if err != nil {
			log.Warnf("Failed to dispatch transaction with url ", d.clientUrls[i], " err = ", err)
			continue
		}

		log.Verbose("Dispatching solana tx successfully signature = ", signature)
		return &types.DispatchedTxResult{
			Success: true,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	}

	return types.NewDispatchTxError(request, types.ErrSubmitTx)
}

// analyzeTx is a function for debugging transaction.
func (d *Dispatcher) analyzeTx(bxBytes []byte) {
	decoder := bin.NewBinDecoder(bxBytes)
	decodedTx := solanago.Transaction{}
	err := decodedTx.UnmarshalWithDecoder(decoder)
	if err != nil {
		log.Error("Failed to decode tx, err = ")
		return
	}

	decodedTx.EncodeTree(text.NewTreeEncoder(os.Stdout, text.Bold("TEST TRANSACTION")))
}

func (d *Dispatcher) sendTransaction(tx []byte, client jsonrpc.RPCClient) (solanago.Signature, error) {
	encodedTx := base64.StdEncoding.EncodeToString(tx)
	opts := rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: "",
	}

	obj := opts.ToMap()
	params := []interface{}{
		encodedTx,
		obj,
	}

	response, err := client.Call(context.Background(), "sendTransaction", params...)
	if err != nil {
		return solanago.Signature{}, err
	}

	if response.Error != nil {
		return solanago.Signature{}, response.Error
	}

	var signature solanago.Signature
	err = response.GetObject(&signature)

	return signature, err
}

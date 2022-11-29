package solana

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	bin "github.com/gagliardetto/binary"
	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/text"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type Dispatcher struct {
	clientUrls []string
	wsUrls     []string

	clients  []*rpc.Client
	wsClient []*ws.Client
}

func NewDispatcher(clientUrls, wsUrls []string) *Dispatcher {
	clients := make([]*rpc.Client, 0)
	wsClients := make([]*ws.Client, 0)

	fmt.Println("clientUrls = ", clientUrls)
	fmt.Println("wsUrls = ", wsUrls)

	for i := range clientUrls {
		client := rpc.New(clientUrls[i])
		wsClient, err := ws.Connect(context.Background(), wsUrls[i])
		if err != nil {
			log.Errorf("Failed to connecto ws client", wsUrls[i])
			continue
		}

		clients = append(clients, client)
		wsClients = append(wsClients, wsClient)
	}
	return &Dispatcher{
		clientUrls: clientUrls,
		wsUrls:     wsUrls,
		clients:    clients,
		wsClient:   wsClients,
	}
}

func (d *Dispatcher) Start() {
}

func (d *Dispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	decoder := bin.NewBinDecoder(request.Tx)
	decodedTx := solanago.Transaction{}
	err := decodedTx.UnmarshalWithDecoder(decoder)
	if err != nil {
		log.Error("Failed to decode tx, err = ", err)
		return &types.DispatchedTxResult{
			Success: false,
			Err:     types.ErrGeneric,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	}

	for i := range d.clients {
		signature, err := d.clients[i].SendEncodedTransactionWithOpts(
			context.Background(),
			base64.StdEncoding.EncodeToString(request.Tx),
			rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: "",
			},
		)
		if err != nil {
			log.Warnf("Failed to dispatch transaction with url ", d.clientUrls[i])
			continue
		}

		log.Verbose("Dispatching solana tx successfully signature = ", signature)
		return &types.DispatchedTxResult{
			Success: true,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	}

	return nil
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

package solana

import (
	"context"
	"encoding/base64"

	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type Dispatcher struct {
	clients  []*rpc.Client
	wsClient []*ws.Client
}

func NewDispatcher(clientUrls, wsUrls []string) *Dispatcher {
	clients := make([]*rpc.Client, 0)
	wsClients := make([]*ws.Client, 0)
	for i := range clientUrls {
		client := rpc.New(clientUrls[i])
		wsClient, err := ws.Connect(context.Background(), wsUrls[i])
		if err != nil {
			continue
		}

		clients = append(clients, client)
		wsClients = append(wsClients, wsClient)
	}
	return &Dispatcher{
		clients:  clients,
		wsClient: wsClients,
	}
}

func (d *Dispatcher) Start() {
}

func (d *Dispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	for i := range d.clients {
		ctx := context.Background()
		signature, err := d.clients[i].SendEncodedTransactionWithOpts(
			ctx,
			base64.StdEncoding.EncodeToString(request.Tx),
			rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
			},
		)
		if err != nil {
			continue
		}

		_, err = confirm.WaitForConfirmation(
			ctx,
			d.wsClient[i],
			signature,
			nil,
		)
		if err == nil {
			log.Verbose("Solana transaction is dispatched successfully")
			break
		}
	}

	return nil
}

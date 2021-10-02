package chains

import (
	"context"
	"fmt"

	eTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

type Dispatcher interface {
	Start()
	Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult
}

type EthDispatcher struct {
	chain, rpcEndpoint string
	client             *ethclient.Client
}

func NewEhtDispatcher(chain, rpcEndpoint string) Dispatcher {
	return &EthDispatcher{
		chain:       chain,
		rpcEndpoint: rpcEndpoint,
	}
}

func (d *EthDispatcher) Start() {
	var err error
	d.client, err = ethclient.Dial(d.rpcEndpoint)
	if err != nil {
		utils.LogError("Cannot dial chain", d.chain, "at endpoint", d.rpcEndpoint)
		// TODO: Add retry mechanism here.
		return
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
	if request.IsEthContractDeployment {
		if request.PubKey == nil {
			return types.NewDispatchTxError(fmt.Errorf("invalid pubkey"))
		}

		from := utils.PublicKeyBytesToAddress(request.PubKey)
		addr = crypto.CreateAddress(from, tx.Nonce()).String()

		utils.LogDebug("Deployed address = ", addr)
	}

	if err := d.client.SendTransaction(context.Background(), tx); err != nil {
		return types.NewDispatchTxError(err)
	}

	return &types.DispatchedTxResult{
		Success:      true,
		DeployedAddr: addr,
	}
}
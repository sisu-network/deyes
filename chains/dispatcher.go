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
	from := utils.PublicKeyBytesToAddress(request.PubKey)
	if request.IsEthContractDeployment {
		if request.PubKey == nil {
			return types.NewDispatchTxError(fmt.Errorf("invalid pubkey"))
		}

		addr = crypto.CreateAddress(from, tx.Nonce()).String()

		utils.LogInfo("Deployed address = ", addr, "for chain", request.Chain)
	}

	if err := d.client.SendTransaction(context.Background(), tx); err != nil {
		// It is possible that another node has deployed the same transaction. We check if the tx has
		// been inclued into the blockchain or not.
		_, _, err2 := d.client.TransactionByHash(context.Background(), tx.Hash())
		if err2 != nil {
			utils.LogError("cannot dispatch tx, from =", from, "chain = ", request.Chain)
			utils.LogError("cannot dispatch tx, err =", err)
			utils.LogError("cannot dispatch tx, err2 =", err2)
			return types.NewDispatchTxError(err)
		}

		utils.LogInfo("The transaction has been deployed before. Tx hash = ", tx.Hash().String())
	}

	utils.LogVerbose("Tx is dispatched successfully for chain", request.Chain, "from", from, "txHash =", tx.Hash())

	return &types.DispatchedTxResult{
		Success:                 true,
		DeployedAddr:            addr,
		Chain:                   request.Chain,
		TxHash:                  request.TxHash,
		IsEthContractDeployment: request.IsEthContractDeployment,
	}
}

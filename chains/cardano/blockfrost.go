package cardano

import (
	"context"
	"fmt"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
)

type blockfrostProvider struct {
	inner blockfrost.APIClient
}

func NewBlockfrostProvider(cfg config.Chain) Provider {
	var chain string
	switch cfg.Chain {
	case "cardano-testnet":
		chain = "preprod"
	default:
		panic("Unknown chain: " + cfg.Chain)
	}

	return &blockfrostProvider{
		inner: blockfrost.NewAPIClient(blockfrost.APIClientOptions{
			ProjectID: cfg.RpcSecret,
			Server:    fmt.Sprintf("https://cardano-%s.blockfrost.io/api/v0", chain),
		}),
	}
}

func (b blockfrostProvider) Health(ctx context.Context) (bool, error) {
	health, err := b.inner.Health(context.Background())
	if err != nil {
		return false, err
	}

	return health.IsHealthy, nil
}

func (b blockfrostProvider) BlockLatest(ctx context.Context) (*types.CardanoBlock, error) {
	block, err := b.inner.BlockLatest(ctx)
	if err != nil {
		return nil, err
	}

	return &types.CardanoBlock{
		Height: block.Height,
		Time:   block.Time,
		Hash:   block.Hash,
	}, nil
}

func (b blockfrostProvider) Block(ctx context.Context, hashOrNumber string) (*types.CardanoBlock, error) {
	block, err := b.inner.Block(ctx, hashOrNumber)
	if err != nil {
		return nil, err
	}

	return &types.CardanoBlock{
		Height: block.Height,
		Time:   block.Time,
		Hash:   block.Hash,
	}, nil
}

func (b blockfrostProvider) AddressTransactions(ctx context.Context, address string, params types.APIQueryParams) ([]*types.AddressTransactions, error) {
	btxs, err := b.inner.AddressTransactions(ctx, address, blockfrost.APIQueryParams{
		Count: params.Count,
		Page:  params.Page,
		Order: params.Order,
		From:  params.From,
		To:    params.To,
	})
	if err != nil {
		return nil, err
	}

	txs := make([]*types.AddressTransactions, 0)
	for _, btx := range btxs {
		txs = append(txs, &types.AddressTransactions{
			TxHash: btx.TxHash,
		})
	}

	return txs, nil
}

func (b blockfrostProvider) TransactionMetadata(ctx context.Context, hash string) ([]*types.TransactionMetadata, error) {
	bmetas, err := b.inner.TransactionMetadata(ctx, hash)
	if err != nil {
		return nil, err
	}

	metas := make([]*types.TransactionMetadata, 0)
	for _, bmeta := range bmetas {
		metas = append(metas, &types.TransactionMetadata{
			JsonMetadata: bmeta.JsonMetadata,
			Label:        bmeta.Label,
		})
	}

	return metas, nil
}

func (b blockfrostProvider) TransactionUTXOs(ctx context.Context, hash string) (*types.TransactionUTXOs, error) {
	transactionUTXOs, err := b.inner.TransactionUTXOs(ctx, hash)
	if err != nil {
		return nil, err
	}

	outputs := make([]types.TransactionUTXOsOutput, 0)
	for _, bOuptut := range transactionUTXOs.Outputs {
		ourOutput := types.TransactionUTXOsOutput{
			Address: bOuptut.Address,
			Amount:  make([]types.TxAmount, len(bOuptut.Amount)),
		}

		for i, amount := range bOuptut.Amount {
			ourOutput.Amount[i] = types.TxAmount{
				Quantity: amount.Quantity,
				Unit:     amount.Unit,
			}
		}

		outputs = append(outputs, ourOutput)
	}

	return &types.TransactionUTXOs{
		Hash: transactionUTXOs.Hash,
	}, nil
}

func (b blockfrostProvider) BlockTransactions(ctx context.Context, height string) ([]string, error) {
	txs, err := b.inner.BlockTransactions(ctx, height)
	if err != nil {
		return nil, err
	}

	arrs := make([]string, 0, len(txs))
	for _, tx := range txs {
		arrs = append(arrs, string(tx))
	}

	return arrs, nil
}

func (b blockfrostProvider) LatestEpochParameters(ctx context.Context) (types.EpochParameters, error) {
	bparams, err := b.inner.LatestEpochParameters(ctx)
	if err != nil {
		return types.EpochParameters{}, err
	}

	return types.EpochParameters{
		Epoch:              bparams.Epoch,
		KeyDeposit:         bparams.KeyDeposit,
		MaxBlockHeaderSize: bparams.MaxBlockHeaderSize,
		MaxBlockSize:       bparams.MaxBlockSize,
		MaxTxSize:          bparams.MaxTxSize,
		MinFeeA:            bparams.MinFeeA,
		MinFeeB:            bparams.MinFeeB,
		MinUtxo:            bparams.MinUtxo,
		NOpt:               bparams.NOpt,
		PoolDeposit:        bparams.PoolDeposit,
	}, nil
}

func (b blockfrostProvider) AddressUTXOs(ctx context.Context, address string, params types.APIQueryParams) ([]types.AddressUTXO, error) {
	butxos, err := b.inner.AddressUTXOs(ctx, address, blockfrost.APIQueryParams{
		Count: params.Count,
		Page:  params.Page,
		Order: params.Order,
		From:  params.From,
		To:    params.To,
	})
	if err != nil {
		return nil, err
	}

	utxos := make([]types.AddressUTXO, len(butxos))

	for i, butxo := range butxos {
		utxos[i] = types.AddressUTXO{
			TxHash:      butxo.TxHash,
			Block:       butxo.Block,
			OutputIndex: butxo.OutputIndex,
			Amount:      make([]types.AddressAmount, 0, len(butxo.Amount)),
		}

		for _, amount := range butxo.Amount {
			utxos[i].Amount = append(utxos[i].Amount, types.AddressAmount{
				Quantity: amount.Quantity,
				Unit:     amount.Unit,
			})
		}
	}

	return nil, nil
}

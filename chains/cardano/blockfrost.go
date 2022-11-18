package cardano

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
	"github.com/sisu-network/deyes/config"
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

func (b blockfrostProvider) BlockLatest(ctx context.Context) (*providertypes.Block, error) {
	block, err := b.inner.BlockLatest(ctx)
	if err != nil {
		return nil, err
	}

	return &providertypes.Block{
		Height: block.Height,
		Time:   block.Time,
		Hash:   block.Hash,
	}, nil
}

func (b blockfrostProvider) Block(ctx context.Context, hashOrNumber string) (*providertypes.Block, error) {
	block, err := b.inner.Block(ctx, hashOrNumber)
	if err != nil {
		return nil, err
	}

	return &providertypes.Block{
		Height: block.Height,
		Time:   block.Time,
		Hash:   block.Hash,
	}, nil
}

func (b blockfrostProvider) AddressTransactions(ctx context.Context, address string, params providertypes.APIQueryParams) ([]*providertypes.AddressTransactions, error) {
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

	txs := make([]*providertypes.AddressTransactions, 0)
	for _, btx := range btxs {
		txs = append(txs, &providertypes.AddressTransactions{
			TxHash: btx.TxHash,
		})
	}

	return txs, nil
}

func (b blockfrostProvider) TransactionMetadata(ctx context.Context, hash string) ([]*providertypes.TransactionMetadata, error) {
	bmetas, err := b.inner.TransactionMetadata(ctx, hash)
	if err != nil {
		return nil, err
	}

	metas := make([]*providertypes.TransactionMetadata, 0)
	for _, bmeta := range bmetas {
		metas = append(metas, &providertypes.TransactionMetadata{
			JsonMetadata: bmeta.JsonMetadata,
			Label:        bmeta.Label,
		})
	}

	return metas, nil
}

func (b blockfrostProvider) TransactionUTXOs(ctx context.Context, hash string) (*providertypes.TransactionUTXOs, error) {
	transactionUTXOs, err := b.inner.TransactionUTXOs(ctx, hash)
	if err != nil {
		return nil, err
	}

	outputs := make([]providertypes.TransactionUTXOsOutput, 0)
	for _, bOuptut := range transactionUTXOs.Outputs {
		ourOutput := providertypes.TransactionUTXOsOutput{
			Address: bOuptut.Address,
			Amount:  make([]providertypes.TxAmount, len(bOuptut.Amount)),
		}

		for i, amount := range bOuptut.Amount {
			ourOutput.Amount[i] = providertypes.TxAmount{
				Quantity: amount.Quantity,
				Unit:     amount.Unit,
			}
		}

		outputs = append(outputs, ourOutput)
	}

	return &providertypes.TransactionUTXOs{
		Hash:    transactionUTXOs.Hash,
		Outputs: outputs,
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

func (b blockfrostProvider) LatestEpochParameters(ctx context.Context) (*cardano.ProtocolParams, error) {
	eparams, err := b.inner.LatestEpochParameters(context.Background())
	if err != nil {
		return nil, err
	}

	minUTXO, err := strconv.ParseUint(eparams.MinUtxo, 10, 64)
	if err != nil {
		return nil, err
	}

	poolDeposit, err := strconv.ParseUint(eparams.PoolDeposit, 10, 64)
	if err != nil {
		return nil, err
	}
	keyDeposit, err := strconv.ParseUint(eparams.KeyDeposit, 10, 64)
	if err != nil {
		return nil, err
	}

	pparams := &cardano.ProtocolParams{
		MinFeeA:            cardano.Coin(eparams.MinFeeA),
		MinFeeB:            cardano.Coin(eparams.MinFeeB),
		MaxBlockBodySize:   uint(eparams.MaxBlockSize),
		MaxTxSize:          uint(eparams.MaxTxSize),
		MaxBlockHeaderSize: uint(eparams.MaxBlockHeaderSize),
		KeyDeposit:         cardano.Coin(keyDeposit),
		PoolDeposit:        cardano.Coin(poolDeposit),
		MaxEpoch:           uint(eparams.Epoch),
		NOpt:               uint(eparams.NOpt),
		CoinsPerUTXOWord:   cardano.Coin(minUTXO),
	}

	return pparams, nil
}

func (b blockfrostProvider) AddressUTXOs(ctx context.Context, address string, params providertypes.APIQueryParams) ([]cardano.UTxO, error) {
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

	utxos := make([]cardano.UTxO, len(butxos))
	spender, err := cardano.NewAddress(address)
	if err != nil {
		return nil, err
	}

	for i, butxo := range butxos {
		fmt.Printf("butxo = %v+\n", butxo)
		txHash, err := cardano.NewHash32(butxo.TxHash)
		if err != nil {
			return nil, err
		}

		amount := cardano.NewValue(0)
		for _, a := range butxo.Amount {
			if a.Unit == "lovelace" {
				lovelace, err := strconv.ParseUint(a.Quantity, 10, 64)
				if err != nil {
					return nil, err
				}
				amount.Coin += cardano.Coin(lovelace)
			} else {
				unitBytes, err := hex.DecodeString(a.Unit)
				if err != nil {
					return nil, err
				}
				policyID := cardano.NewPolicyIDFromHash(unitBytes[:28])
				assetName := string(unitBytes[28:])
				assetValue, err := strconv.ParseUint(a.Quantity, 10, 64)
				if err != nil {
					return nil, err
				}
				currentAssets := amount.MultiAsset.Get(policyID)
				if currentAssets != nil {
					currentAssets.Set(
						cardano.NewAssetName(assetName),
						cardano.BigNum(assetValue),
					)
				} else {
					amount.MultiAsset.Set(
						policyID,
						cardano.NewAssets().
							Set(
								cardano.NewAssetName(string(assetName)),
								cardano.BigNum(assetValue),
							),
					)
				}
			}
		}

		utxos[i] = cardano.UTxO{
			Spender: spender,
			TxHash:  txHash,
			Amount:  amount,
			Index:   uint64(butxo.OutputIndex),
		}
	}

	return utxos, nil
}

func (b blockfrostProvider) Tip(blockHeight uint64) (*cardano.NodeTip, error) {
	block, err := b.inner.Block(context.Background(), fmt.Sprintf("%d", blockHeight))
	if err != nil {
		return nil, err
	}

	return &cardano.NodeTip{
		Block: uint64(block.Height),
		Epoch: uint64(block.Epoch),
		Slot:  uint64(block.Slot),
	}, nil
}

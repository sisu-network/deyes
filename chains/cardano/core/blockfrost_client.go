package core

import (
	"context"
	"strconv"
	"sync"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/lib/log"
)

const (
	ParamsOrderDesc = "desc"
	UnitLovelace    = "lovelace"
)

// implements cardanoClient
type BlockfrostClient2 struct {
	inner blockfrost.APIClient

	// cache assets
	policyAssets map[string]*cardano.Assets
	bfAssetCache map[string]*blockfrost.Asset
	lock         *sync.RWMutex
}

func NewBlockfrostClient2(options blockfrost.APIClientOptions) CardanoClient {
	return &BlockfrostClient2{
		inner:        blockfrost.NewAPIClient(options),
		policyAssets: make(map[string]*cardano.Assets),
		bfAssetCache: make(map[string]*blockfrost.Asset),
		lock:         &sync.RWMutex{},
	}
}

func (b *BlockfrostClient2) GetBlockHeight() (int, error) {
	block, err := b.inner.BlockLatest(context.Background())
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

func (b *BlockfrostClient2) GetNewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*cardano.Tx, error) {
	txs := make([]*cardano.Tx, 0)
	latestHeight, err := b.GetBlockHeight()
	if err != nil {
		return nil, err
	}

	if latestHeight < fromHeight {
		return nil, BlockNotFound
	}

	added := make(map[string]bool)

	for addr := range interestedAddrs {
		bfTxs, err := b.inner.AddressTransactions(context.Background(), addr, blockfrost.APIQueryParams{
			Order: ParamsOrderDesc,
			From:  strconv.Itoa(fromHeight),
			To:    strconv.Itoa(fromHeight),
		})

		if err != nil {
			return nil, err
		}

		for _, bfTx := range bfTxs {
			if added[bfTx.TxHash] {
				continue
			}
			added[bfTx.TxHash] = true

			bfTransactionContent, err := b.inner.Transaction(context.Background(), bfTx.TxHash)
			if err != nil {
				return nil, err
			}

			txUtxos, err := b.inner.TransactionUTXOs(context.Background(), bfTx.TxHash)
			if err != nil {
				return nil, err
			}

			fee, err := strconv.Atoi(bfTransactionContent.Fees)
			if err != nil {
				return nil, err
			}

			tx := &cardano.Tx{
				Body: cardano.TxBody{
					Inputs:  make([]*cardano.TxInput, 0),
					Outputs: make([]*cardano.TxOutput, 0),
					Fee:     cardano.Coin(fee),
				},
			}

			// Recreate input
			for _, input := range txUtxos.Inputs {
				value, err := b.blockfrostAmountToMultiAsset(input.Amount)
				if err != nil {
					return nil, err
				}

				tx.Body.Inputs = append(tx.Body.Inputs, &cardano.TxInput{
					TxHash: cardano.Hash32(input.TxHash),
					Index:  uint64(input.OutputIndex),
					Amount: value,
				})
			}

			// Recreate output
			for _, output := range txUtxos.Outputs {
				addr, err := cardano.NewAddress(output.Address)
				if err != nil {
					return nil, err
				}

				value, err := b.blockfrostAmountToMultiAsset(output.Amount)
				if err != nil {
					return nil, err
				}

				tx.Body.Outputs = append(tx.Body.Outputs, &cardano.TxOutput{
					Address: addr,
					Amount:  value,
				})
			}

			txs = append(txs, tx)
		}
	}

	return txs, nil
}

func (b *BlockfrostClient2) blockfrostAmountToMultiAsset(amounts []blockfrost.TxAmount) (*cardano.Value, error) {
	if len(amounts) == 1 {
		quantity, err := strconv.Atoi(amounts[0].Quantity)
		if err != nil {
			return nil, err
		}

		return cardano.NewValue(cardano.Coin(quantity)), nil
	}

	var coin cardano.Coin
	multiAssets := cardano.NewMultiAsset()

	for _, amount := range amounts {
		if amount.Unit == UnitLovelace {
			quantity, err := strconv.Atoi(amount.Quantity)
			if err != nil {
				return nil, err
			}

			coin = cardano.Coin(quantity)
			continue
		}

		bfAsset := b.getCachedAsset(amount.Unit)
		assets, err := b.getAssetsFromPolicy(bfAsset.PolicyId)
		if err != nil {
			return nil, err
		}

		multiAssets.Set(cardano.NewPolicyIDFromHash(cardano.Hash28(amount.Unit)), assets)
	}

	return cardano.NewValueWithAssets(coin, multiAssets), nil
}

func (b *BlockfrostClient2) getAssetsFromPolicy(policyHash string) (*cardano.Assets, error) {
	var assets *cardano.Assets
	b.lock.RLock()
	assets = b.policyAssets[policyHash]
	b.lock.RUnlock()

	if assets != nil {
		return assets, nil
	}

	bfAssets, err := b.inner.AssetsByPolicy(context.Background(), policyHash)
	if err != nil {
		return nil, err
	}

	assets = cardano.NewAssets()
	for _, bfAsset := range bfAssets {
		val, err := strconv.Atoi(bfAsset.Quantity)
		if err != nil {
			log.Error(err)
			val = 0
		}

		assets.Set(cardano.NewAssetName(bfAsset.Asset), cardano.BigNum(val))
	}

	return assets, nil
}

func (b *BlockfrostClient2) getCachedAsset(assetId string) *blockfrost.Asset {
	b.lock.RLock()
	asset := b.bfAssetCache[assetId]
	b.lock.RUnlock()

	if asset != nil {
		return asset
	}

	a, err := b.inner.Asset(context.Background(), assetId)
	if err != nil {
		return nil
	}
	asset = &a

	b.lock.Lock()
	b.bfAssetCache[assetId] = asset
	b.lock.Unlock()

	return asset
}

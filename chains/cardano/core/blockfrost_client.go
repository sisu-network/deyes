package core

import (
	"context"
	"encoding/hex"
	"strconv"
	"sync"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

const (
	ParamsOrderDesc = "desc"
	UnitLovelace    = "lovelace"
)

// implements cardanoClient
type BlockfrostClient struct {
	inner blockfrost.APIClient

	// cache assets
	policyAssets map[string]*cardano.Assets
	bfAssetCache map[string]*blockfrost.Asset
	lock         *sync.RWMutex
}

func NewBlockfrostClient(options blockfrost.APIClientOptions) CardanoClient {
	return &BlockfrostClient{
		inner:        blockfrost.NewAPIClient(options),
		policyAssets: make(map[string]*cardano.Assets),
		bfAssetCache: make(map[string]*blockfrost.Asset),
		lock:         &sync.RWMutex{},
	}
}

func (b *BlockfrostClient) IsHealthy() bool {
	health, err := b.inner.Health(context.Background())
	if err != nil {
		log.Error("Blockfrost is not healthy")
		return false
	}

	return health.IsHealthy
}

func (b *BlockfrostClient) LatestBlock() *blockfrost.Block {
	block, err := b.inner.BlockLatest(context.Background())
	if err != nil {
		log.Error("Failed to get latest cardano block, err = ", err)
		return nil
	}

	return &block
}

func (b *BlockfrostClient) BlockHeight() (int, error) {
	block, err := b.inner.BlockLatest(context.Background())
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

func (b *BlockfrostClient) NewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*types.CardanoUtxo, error) {
	latestHeight, err := b.BlockHeight()
	if err != nil {
		return nil, err
	}

	if latestHeight < fromHeight {
		return nil, BlockNotFound
	}

	added := make(map[string]bool)
	utxos := make([]*types.CardanoUtxo, 0)

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
			txUtxos, err := b.inner.TransactionUTXOs(context.Background(), bfTx.TxHash)
			if err != nil {
				return nil, err
			}

			for i, utxo := range txUtxos.Outputs {
				cardanoAddr, err := cardano.NewAddress(addr)
				if err != nil {
					return nil, err
				}

				amount, err := b.getCardanoAmount(utxo.Amount)
				if err != nil {
					return nil, err
				}

				cardanoUtxo := &types.CardanoUtxo{
					Spender: cardanoAddr,
					TxHash:  cardano.Hash32(txUtxos.Hash),
					Amount:  amount,
					Index:   uint64(i),
				}
				utxos = append(utxos, cardanoUtxo)
			}
		}
	}

	return utxos, nil
}

func (b *BlockfrostClient) getCardanoAmount(amounts []blockfrost.TxAmount) (*cardano.Value, error) {
	amount := cardano.NewValue(0)
	for _, a := range amounts {
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

	return amount, nil
}

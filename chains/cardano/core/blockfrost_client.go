package core

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
	inner   blockfrost.APIClient
	options blockfrost.APIClientOptions

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

// IsHealthy implements CardanoClient
func (b *BlockfrostClient) IsHealthy() bool {
	health, err := b.inner.Health(context.Background())
	if err != nil {
		log.Error("Blockfrost is not healthy")
		return false
	}

	return health.IsHealthy
}

// LatestBlock implements CardanoClient
func (b *BlockfrostClient) LatestBlock() *blockfrost.Block {
	block, err := b.inner.BlockLatest(context.Background())
	if err != nil {
		log.Error("Failed to get latest cardano block, err = ", err)
		return nil
	}

	return &block
}

// BlockHeight implements CardanoClient
func (b *BlockfrostClient) BlockHeight() (int, error) {
	block, err := b.inner.BlockLatest(context.Background())
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

// NewTxs implements CardanoClient
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

// SubmitTx implements CardanoClient
func (b *BlockfrostClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	// Copy from this https://github.com/echovl/cardano-go/blob/4936c872fbb1f1db4bf04f1242fc180b0fe9843f/blockfrost/blockfrost.go#L124
	url := fmt.Sprintf("%s/tx/submit", b.options.Server)
	txBytes := tx.Bytes()

	req, err := http.NewRequest("POST", url, bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("project_id", b.options.ProjectID)
	req.Header.Add("Content-Type", "application/cbor")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

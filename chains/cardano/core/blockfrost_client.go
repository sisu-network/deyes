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
	"github.com/sisu-network/deyes/chains/cardano/utils"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

var _ CardanoClient = (*BlockfrostClient)(nil)

const (
	ParamsOrderDesc = "desc"
	UnitLovelace    = "lovelace"
)

var (
	MetadataNotFound = fmt.Errorf("Metadata not found")
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

func NewBlockfrostClient(options blockfrost.APIClientOptions) *BlockfrostClient {
	return &BlockfrostClient{
		options:      options,
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
	block, err := b.inner.BlockLatest(b.getContext())
	if err != nil {
		log.Error("Failed to get latest cardano block, err = ", err)
		return nil
	}

	return &block
}

func (b *BlockfrostClient) GetBlock(hashOrNumber string) (*blockfrost.Block, error) {
	block, err := b.inner.Block(b.getContext(), hashOrNumber)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// BlockHeight implements CardanoClient
func (b *BlockfrostClient) BlockHeight() (int, error) {
	block, err := b.inner.BlockLatest(b.getContext())
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

// NewTxs implements CardanoClient
func (b *BlockfrostClient) NewTxs(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error) {
	latestHeight, err := b.BlockHeight()
	if err != nil {
		return nil, err
	}

	if latestHeight < fromHeight {
		return nil, BlockNotFound
	}

	txs := make([]*types.CardanoTransactionUtxo, 0)

	txHahes, err := b.inner.BlockTransactions(context.Background(), fmt.Sprintf("%d", fromHeight))
	if err != nil {
		return nil, err
	}

	for _, txHash := range txHahes {
		utxos, err := b.inner.TransactionUTXOs(context.Background(), string(txHash))
		if err != nil {
			return nil, err
		}

		if !b.shouldIncludeTx(utxos, gateway) {
			continue
		}

		metadata, err := b.GetTransactionMetadata(string(txHash))
		if err != nil && err != MetadataNotFound {
			return nil, err
		}

		for i, output := range utxos.Outputs {
			if output.Address == gateway {
				tx := &types.CardanoTransactionUtxo{
					Hash:     string(txHash),
					Index:    i,
					Address:  output.Address,
					Metadata: metadata,
					Amount:   make([]types.TxAmount, len(output.Amount)),
				}

				for j, amount := range output.Amount {
					tx.Amount[j] = types.TxAmount{
						Quantity: amount.Quantity,
						Unit:     amount.Unit,
					}
				}

				txs = append(txs, tx)
			}
		}
	}

	return txs, nil
}

func (b *BlockfrostClient) shouldIncludeTx(utxos blockfrost.TransactionUTXOs, gateway string) bool {
	for _, output := range utxos.Outputs {
		if output.Address == gateway {
			return true
		}
	}

	return false
}

func (b *BlockfrostClient) getCardanoAmount(amounts []blockfrost.TxAmount) (*cardano.Value, error) {
	amount := cardano.NewValue(0)
	for _, a := range amounts {
		if a.Unit == "lovelace" {
			lovelace, err := strconv.ParseUint(a.Quantity, 10, 64)
			if err != nil {
				log.Error("error when parsing lovelace unit: ", err)
				return nil, err
			}
			amount.Coin += cardano.Coin(lovelace)
		} else {
			unitBytes, err := hex.DecodeString(a.Unit)
			if err != nil {
				log.Error("error when decode multi-asset unit: ", err)
				return nil, err
			}
			policyID := cardano.NewPolicyIDFromHash(unitBytes[:28])
			assetName := string(unitBytes[28:])
			assetValue, err := strconv.ParseUint(a.Quantity, 10, 64)
			if err != nil {
				log.Error("error when parsing multi-asset value: ", err)
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

func (b *BlockfrostClient) GetTransactionMetadata(txHash string) (*types.CardanoTxMetadata, error) {
	txMetadata, err := b.inner.TransactionMetadata(b.getContext(), txHash)
	if err != nil {
		log.Error("error when getting transaction metadata: ", err)
		return nil, err
	}

	if len(txMetadata) == 0 {
		return nil, MetadataNotFound
	}

	// Noted: when creating a transaction with metadata, please attach metadata in label "0"
	log.Debug("Label = ", txMetadata[0].Label)
	txMetadatum, ok := txMetadata[0].JsonMetadata.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("unknown tx metadatum type. Expected map[string]interface{}, got: %T", txMetadata[0].JsonMetadata)
		log.Error(err)
		return nil, err
	}

	txAdditionInfo := &types.CardanoTxMetadata{}
	if err := utils.MapToJSONStruct(txMetadatum, txAdditionInfo); err != nil {
		log.Error(err)
		return nil, err
	}

	return txAdditionInfo, nil
}

func (b *BlockfrostClient) getContext() context.Context {
	return context.Background()
}

// SubmitTx implements CardanoClient
func (b *BlockfrostClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	for _, i := range tx.Body.Inputs {
		log.Debugf("tx input = %+v\n", i)
	}

	for _, o := range tx.Body.Outputs {
		log.Debugf("tx output = %+v\n", o)
	}

	log.Debugf("tx = %+v\n", tx)
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

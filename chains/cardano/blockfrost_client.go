package cardano

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/chains/cardano/utils"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

type CardanoClient interface {
	IsHealthy() bool
	LatestBlock() (*blockfrost.Block, error)
	GetBlock(hashOrNumber string) (*blockfrost.Block, error)
	BlockHeight() (int, error)
	NewTxs(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error)
	SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error)
}

var _ CardanoClient = (*DefaultCardanoClient)(nil)

type Provider interface {
	Health(ctx context.Context) (blockfrost.Health, error)
	BlockLatest(ctx context.Context) (blockfrost.Block, error)
	Block(ctx context.Context, hashOrNumber string) (blockfrost.Block, error)
	AddressTransactions(ctx context.Context, address string, query blockfrost.APIQueryParams) ([]blockfrost.AddressTransactions, error)
	TransactionMetadata(ctx context.Context, hash string) ([]blockfrost.TransactionMetadata, error)
	TransactionUTXOs(ctx context.Context, hash string) (blockfrost.TransactionUTXOs, error)
	BlockTransactions(ctx context.Context, height string) ([]blockfrost.Transaction, error)
	LatestEpochParameters(ctx context.Context) (blockfrost.EpochParameters, error)
	AddressUTXOs(ctx context.Context, address string, query blockfrost.APIQueryParams) ([]blockfrost.AddressUTXO, error)
}

const (
	ParamsOrderDesc = "desc"
	UnitLovelace    = "lovelace"
)

var (
	MetadataNotFound = fmt.Errorf("Metadata not found")
)

// DefaultCardanoClient implements CardanoClient
type DefaultCardanoClient struct {
	inner       Provider
	submitTxURL string
	secret      string

	// cache assets
	policyAssets map[string]*cardano.Assets
	bfAssetCache map[string]*blockfrost.Asset
	lock         *sync.RWMutex
}

func NewDefaultCardanoClient(inner Provider, submitTxURL, secret string) *DefaultCardanoClient {
	return &DefaultCardanoClient{
		inner:        inner,
		secret:       secret,
		submitTxURL:  submitTxURL,
		policyAssets: make(map[string]*cardano.Assets),
		bfAssetCache: make(map[string]*blockfrost.Asset),
		lock:         &sync.RWMutex{},
	}
}

// IsHealthy implements CardanoClient
func (b *DefaultCardanoClient) IsHealthy() bool {
	health, err := b.inner.Health(context.Background())
	if err != nil {
		log.Error("Blockfrost is not healthy")
		return false
	}

	return health.IsHealthy
}

// LatestBlock implements CardanoClient
func (b *DefaultCardanoClient) LatestBlock() (*blockfrost.Block, error) {
	block, err := b.inner.BlockLatest(b.getContext())
	if err != nil {
		log.Error("Failed to get latest cardano block, err = ", err)
		return nil, err
	}

	return &block, nil
}

func (b *DefaultCardanoClient) GetBlock(hashOrNumber string) (*blockfrost.Block, error) {
	block, err := b.inner.Block(b.getContext(), hashOrNumber)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

// BlockHeight implements CardanoClient
func (b *DefaultCardanoClient) BlockHeight() (int, error) {
	block, err := b.inner.BlockLatest(b.getContext())
	if err != nil {
		return 0, err
	}

	return block.Height, nil
}

// NewTxs implements CardanoClient
func (b *DefaultCardanoClient) NewTxs(fromHeight int, gateway string) ([]*types.CardanoTransactionUtxo, error) {
	latestHeight, err := b.BlockHeight()
	if err != nil {
		return nil, err
	}

	if latestHeight < fromHeight {
		return nil, BlockNotFound
	}

	txs := make([]*types.CardanoTransactionUtxo, 0)

	txHashes, err := b.inner.BlockTransactions(context.Background(), fmt.Sprintf("%d", fromHeight))
	if err != nil {
		return nil, err
	}

	for _, txHash := range txHashes {
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

func (b *DefaultCardanoClient) shouldIncludeTx(utxos blockfrost.TransactionUTXOs, gateway string) bool {
	for _, output := range utxos.Outputs {
		if output.Address == gateway {
			return true
		}
	}

	return false
}

func (b *DefaultCardanoClient) GetTransactionMetadata(txHash string) (*types.CardanoTxMetadata, error) {
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

func (b *DefaultCardanoClient) getContext() context.Context {
	return context.Background()
}

// SubmitTx implements CardanoClient
func (b *DefaultCardanoClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	for _, i := range tx.Body.Inputs {
		log.Debugf("tx input = %+v\n", i)
	}

	for _, o := range tx.Body.Outputs {
		log.Debugf("tx output = %+v\n", o)
	}

	log.Debugf("tx = %+v\n", tx)
	// Copy from this https://github.com/echovl/cardano-go/blob/4936c872fbb1f1db4bf04f1242fc180b0fe9843f/blockfrost/blockfrost.go#L124
	txBytes := tx.Bytes()

	url := b.submitTxURL
	req, err := http.NewRequest("POST", url, bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	// This header is only used for Blockfrost.io call
	req.Header.Add("project_id", b.secret)
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, errors.New(string(respBody))
	}

	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

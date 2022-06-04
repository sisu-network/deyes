package core

import (
	"context"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
)

type CardanoClient interface {
	GetBlockHeight() (int, error)
	GetNewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*cardano.Tx, error)
}

// A wrapper around cardano node client so that we can mock in watcher tests.
type blockfrostClient struct {
	inner blockfrost.APIClient
}

func newAPIClient(options blockfrost.APIClientOptions) *blockfrostClient {
	return &blockfrostClient{
		inner: blockfrost.NewAPIClient(options),
	}
}

func (b *blockfrostClient) Health(ctx context.Context) (blockfrost.Health, error) {
	return b.inner.Health(ctx)
}

func (b *blockfrostClient) BlockLatest(ctx context.Context) (blockfrost.Block, error) {
	return b.inner.BlockLatest(context.Background())
}

func (b *blockfrostClient) AddressUTXOs(ctx context.Context, address string,
	query blockfrost.APIQueryParams) ([]blockfrost.AddressUTXO, error) {
	return b.inner.AddressUTXOs(ctx, address, query)
}

func (b *blockfrostClient) AddressDetails(ctx context.Context, address string) (blockfrost.AddressDetails, error) {
	return b.inner.AddressDetails(ctx, address)
}

func (b *blockfrostClient) Transaction(ctx context.Context, hash string) (blockfrost.TransactionContent, error) {

	return b.inner.Transaction(ctx, hash)
}

func (b *blockfrostClient) AddressTransactions(ctx context.Context, address string,
	query blockfrost.APIQueryParams) ([]blockfrost.AddressTransactions, error) {
	return b.inner.AddressTransactions(ctx, address, query)
}

func (b *blockfrostClient) TransactionUTXOs(ctx context.Context, hash string) (blockfrost.TransactionUTXOs, error) {
	return b.inner.TransactionUTXOs(ctx, hash)
}

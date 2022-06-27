package core

import (
	"context"
	"errors"
	"log"

	"github.com/cardano-community/koios-go-client"
)

type KoiosClient struct {
	inner   koios.Client
	options koios.Option
}

func NewKoiosClient(options koios.Option) *KoiosClient {
	inner, err := koios.New(options)

	if err != nil {
		return nil
	}

	return &KoiosClient{
		options: options,
		inner:   *inner,
	}

}

func (k *KoiosClient) GetTip() *koios.TipResponse {
	tip, err := k.inner.GetTip(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return tip
}

func (k *KoiosClient) IsHealthy() bool {
	tip := k.GetTip()

	if tip == nil {
		log.Fatal(errors.New("No tip response"))
		return false
	}

	return tip.StatusCode == 200
}

func (k *KoiosClient) GetBlock(hash koios.BlockHash) *koios.BlockInfoResponse {
	block, err := k.inner.GetBlockInfo(context.Background(), hash, nil)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return block
}

func (k *KoiosClient) LatestBlock() *koios.Block {

	options := k.inner.NewRequestOptions()
	options.Page(1)
	options.PageSize(1)

	blocks, err := k.inner.GetBlocks(context.Background(), options)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &blocks.Data[0]
}

func (k *KoiosClient) BlockHeight() (int, error) {
	tip := k.GetTip()

	if tip == nil {
		return 0, errors.New("No tip response")
	}

	return tip.Data.BlockNo, nil
}

func (k *KoiosClient) NewTxs(fromHeight int, interestedAddrs map[string]bool) {
	// k.inner.GetAddressTxs()

}

func (k *KoiosClient) SubmitTx() {
	// k.inner.SubmitSignedTx()
}

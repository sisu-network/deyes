package sushiswap

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"math/big"
)

type SushiSwapManager interface {
	GetPriceFromSushiswap(tokenAddress string) (*types.TokenPrice, error)
}

type defaultSushiSwapManager struct {
	cfg config.Deyes
}

func NewSushiSwapManager(cfg config.Deyes) SushiSwapManager {
	return &defaultSushiSwapManager{
		cfg: cfg,
	}
}

func (m *defaultSushiSwapManager) GetPriceFromSushiswap(tokenAddress string) (*types.TokenPrice, error) {
	rpcEth := m.cfg.EthRpc
	daiTokenAddress := m.cfg.DaiTokenAddress
	ctx := context.Background()

	ec, _ := ethclient.DialContext(ctx, rpcEth)
	client := NewClient(ec)
	price, err := client.GetExchangeAmount(big.NewInt(1), common.HexToAddress(tokenAddress),
		common.HexToAddress(daiTokenAddress))

	if err != nil {
		return nil, err
	}
	return &types.TokenPrice{
		Id:    tokenAddress,
		Price: price,
	}, nil
}

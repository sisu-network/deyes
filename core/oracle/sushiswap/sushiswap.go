package sushiswap

import (
	"context"
	"github.com/sisu-network/deyes/core/oracle/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
)

type SushiSwapManager interface {
	GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2, tokenName string) (*types.TokenPrice, error)
}

type defaultSushiSwapManager struct {
	cfg config.Deyes
}

func NewSushiSwapManager(cfg config.Deyes) SushiSwapManager {
	return &defaultSushiSwapManager{
		cfg: cfg,
	}
}

func (m *defaultSushiSwapManager) GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2, tokenName string) (*types.TokenPrice, error) {
	ethRpcs := m.cfg.EthRpcs
	ctx := context.Background()
	clients := make([]*ethclient.Client, 0)
	for _, rpc := range ethRpcs {
		ec, _ := ethclient.DialContext(ctx, rpc)
		clients = append(clients, ec)
	}
	return utils.ExecuteWithClients(clients, func(ethClient *ethclient.Client) (*types.TokenPrice, bool, error) {
		client := NewClient(ethClient)
		price, err := client.GetExchangeAmount(big.NewInt(1), common.HexToAddress(tokenAddress1),
			common.HexToAddress(tokenAddress2))

		if err != nil {
			return nil, false, err
		}
		return &types.TokenPrice{
			Id:    tokenName,
			Price: price,
		}, true, nil
	})

}

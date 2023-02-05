package sushiswap

import (
	"context"
	"math/big"

	"github.com/sisu-network/deyes/core/oracle/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
)

type SushiSwapManager interface {
	GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2 string, amount *big.Int) (
		*big.Int, error)
}

type defaultSushiSwapManager struct {
	cfg config.Deyes
}

func NewSushiSwapManager(cfg config.Deyes) SushiSwapManager {
	return &defaultSushiSwapManager{
		cfg: cfg,
	}
}

func (m *defaultSushiSwapManager) GetPriceFromSushiswap(tokenAddress1 string, tokenAddress2 string,
	amount *big.Int) (*big.Int, error) {

	ethRpcs := m.cfg.EthRpcs
	ctx := context.Background()
	clients := make([]*ethclient.Client, 0)
	for _, rpc := range ethRpcs {
		ec, _ := ethclient.DialContext(ctx, rpc)
		clients = append(clients, ec)
	}

	return utils.ExecuteWithClients(clients, func(ethClient *ethclient.Client) (*big.Int,
		bool, error) {
		client := NewClient(ethClient)
		price, err := client.GetExchangeAmount(
			amount,
			common.HexToAddress(tokenAddress1),
			common.HexToAddress(tokenAddress2),
		)

		if err != nil {
			return nil, false, err
		}
		return price, true, nil
	})

}

package uniswap

import (
	"fmt"
	"math/big"

	"github.com/sisu-network/deyes/core/oracle/utils"

	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
)

type UniswapManager interface {
	GetPriceFromUniswap(tokenAddress1 string, tokenAddress2 string) (*big.Int, error)
}

type defaultUniswapManager struct {
	cfg config.Deyes
}

func NewUniwapManager(cfg config.Deyes) UniswapManager {
	return &defaultUniswapManager{
		cfg: cfg,
	}
}

func (m *defaultUniswapManager) GetPriceFromUniswap(tokenAddress1 string, tokenAddress2 string) (
	*big.Int, error) {
	ethRpcs := m.cfg.EthRpcs

	clients := make([]*ethclient.Client, 0)
	for _, rpc := range ethRpcs {
		ec, err := ethclient.Dial(rpc)
		if err != nil {
			return nil, err
		}
		clients = append(clients, ec)
	}
	return utils.ExecuteWithClients(clients, func(client *ethclient.Client) (*big.Int, bool, error) {
		quoterContract, err := contract.NewUniswapv3Quoter(common.HexToAddress(helper.ContractV3Quoter),
			client)
		if err != nil {
			return nil, false, err
		}
		token0 := common.HexToAddress(tokenAddress1)
		token1 := common.HexToAddress(tokenAddress2) // DAI

		fee := big.NewInt(3000)
		amountIn := helper.FloatStringToBigInt("1.00", 18)
		sqrtPriceLimitX96 := big.NewInt(0)

		var out []interface{}
		rawCaller := &contract.Uniswapv3QuoterRaw{Contract: quoterContract}
		err = rawCaller.Call(nil, &out, "quoteExactInputSingle", token0, token1,
			fee, amountIn, sqrtPriceLimitX96)
		if err != nil {
			return nil, false, err
		}

		if len(out) != 1 {
			return nil, false, fmt.Errorf("Output length is not 1")
		}

		price, ok := out[0].(*big.Int)
		if !ok {
			return nil, false, fmt.Errorf("Failed to cast output to price integer. Out = %v", out)
		}

		return price, true, nil
	})
}

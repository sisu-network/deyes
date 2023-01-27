package uniswap

import (
	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"math/big"
)

type UniswapManager interface {
	GetPriceFromUniswap(tokenAddress string) (*types.TokenPrice, error)
}

type defaultUniswapManager struct {
	cfg config.Deyes
}

func NewUniwapManager(cfg config.Deyes) UniswapManager {
	return &defaultUniswapManager{
		cfg: cfg,
	}
}

func (m *defaultUniswapManager) GetPriceFromUniswap(tokenAddress string) (*types.TokenPrice, error) {
	rpcEth := m.cfg.EthRpc
	daiTokenAddress := m.cfg.DaiTokenAddress
	client, err := ethclient.Dial(rpcEth)
	if err != nil {
		return nil, err
	}
	quoterContract, err := contract.NewUniswapv3Quoter(common.HexToAddress(helper.ContractV3Quoter), client)
	if err != nil {
		return nil, err
	}

	token0 := common.HexToAddress(tokenAddress)
	token1 := common.HexToAddress(daiTokenAddress) // DAI
	fee := big.NewInt(3000)
	amountIn := helper.FloatStringToBigInt("1.00", 18)
	sqrtPriceLimitX96 := big.NewInt(0)

	var out []interface{}
	rawCaller := &contract.Uniswapv3QuoterRaw{Contract: quoterContract}
	err = rawCaller.Call(nil, &out, "quoteExactInputSingle", token0, token1,
		fee, amountIn, sqrtPriceLimitX96)
	if err != nil {
		return nil, err
	}

	return &types.TokenPrice{
		Id:    tokenAddress,
		Price: out[0].(*big.Int),
	}, nil
}

package eth

import (
	"context"
	"math/big"
	"testing"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/utils"
	"github.com/stretchr/testify/require"
)

func getTestTxForDynamicGas(baseFee, tip int64) *ethtypes.Transaction {
	feeCap := 2*baseFee + tip
	dynamicFee := &ethtypes.DynamicFeeTx{
		GasTipCap: big.NewInt(tip),
		GasFeeCap: big.NewInt(feeCap),
	}
	return ethtypes.NewTx(dynamicFee)
}

func TestDynamicGasFee(t *testing.T) {
	cfg := config.Chain{
		UseGasEip1559: true,
	}

	header := &ethtypes.Header{
		BaseFee: big.NewInt(10 * 1_000_000_000),
	}
	txs := make([]*ethtypes.Transaction, 0)
	receipts := make([]*ethtypes.Receipt, 0)
	for i := 0; i < GasQueueSize; i++ {
		tx := getTestTxForDynamicGas(header.BaseFee.Int64(), int64(i*1_000_000_000))
		txs = append(txs, tx)
		receipts = append(receipts, &ethtypes.Receipt{})
	}
	block := ethtypes.NewBlock(header, txs, []*ethtypes.Header{}, receipts, trie.NewStackTrie(nil))

	client := &MockEthClient{}
	gasCal := newGasCalculator(cfg, client, GasPriceUpdateInterval)

	gasCal.AddNewBlock(block)

	baseFee := gasCal.GetBaseFee()
	require.Equal(t, big.NewInt(10000000000), baseFee)
	tip := gasCal.GetTip()
	require.Equal(t, big.NewInt(19500000000), tip)
}

func TestLegacyGasPrice(t *testing.T) {
	cfg := config.Chain{}
	client := &MockEthClient{
		SuggestGasPriceFunc: func(ctx context.Context) (*big.Int, error) {
			return big.NewInt(utils.OneGweiInWei * 10), nil
		},
	}

	updateInterval := time.Millisecond * 500
	gasCal := newGasCalculator(cfg, client, updateInterval)
	gasCal.Start()

	require.Equal(t, big.NewInt(utils.OneGweiInWei*10), gasCal.GetGasPrice())
	client.SuggestGasPriceFunc = func(ctx context.Context) (*big.Int, error) {
		return big.NewInt(utils.OneGweiInWei * 12), nil
	}

	// Gas price is still the old price
	require.Equal(t, big.NewInt(utils.OneGweiInWei*10), gasCal.GetGasPrice())
	time.Sleep(updateInterval)

	require.Equal(t, big.NewInt(utils.OneGweiInWei*12), gasCal.GetGasPrice())
}

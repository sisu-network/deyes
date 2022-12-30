package eth

import (
	"context"
	"math/big"
	"sync"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/lib/log"
)

const (
	GasQueueSize    = 40
	DefaultBaseFee  = int64(15_000_000_000) // 15 gwei
	DefaultGasPrice = int64(20_000_000_000) // 20 gwei
	DefaultTip      = int64(1_000_000_000)
)

var (
	GasPriceUpdateInterval = time.Second * 60
)

// gasCalculator is auxiliary struct that calculates gas price (for legacy tx), base fee & tip (
// for EIP 1559 tx) based on recent transactions.
type gasCalculator struct {
	cfg                    config.Chain
	client                 EthClient
	gasPrice               int64
	gasPriceUpdateInterval time.Duration

	lastUpdateGasPrice time.Time
	baseFeeQueue       []int64
	tipQueue           []int64
	queueIndex         int
	lock               *sync.RWMutex
}

func newGasCalculator(cfg config.Chain, client EthClient,
	gasPriceUpdateInterval time.Duration) *gasCalculator {
	return &gasCalculator{
		cfg:          cfg,
		client:       client,
		baseFeeQueue: make([]int64, 0, GasQueueSize),
		tipQueue:     make([]int64, 0, GasQueueSize),
		lock:         &sync.RWMutex{},
	}
}

func (g *gasCalculator) Start() {
	g.updateGasPrice()
}

// AddNewBlock takes as new ETH block and update base fee & tip (for EIP 1559).
func (g *gasCalculator) AddNewBlock(block *ethtypes.Block) {
	if g.cfg.UseGasEip1559 {
		// Update base tip
		for _, tx := range block.Transactions() {
			// Check tx gas base fee
			switch tx.Type() {
			case ethtypes.DynamicFeeTxType:
				g.enqueue(block.BaseFee(), tx.GasTipCap())
			}
		}
	}
}

func (g *gasCalculator) enqueue(baseFee, tip *big.Int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if len(g.baseFeeQueue) < GasQueueSize {
		g.baseFeeQueue = append(g.baseFeeQueue, 0)
		g.tipQueue = append(g.tipQueue, 0)
	}

	next := (g.queueIndex + 1) % len(g.baseFeeQueue)

	g.baseFeeQueue[next] = baseFee.Int64()
	g.tipQueue[next] = tip.Int64()
	g.queueIndex = next
}

// GetBaseFee returns the estimated base fee.
func (g *gasCalculator) GetBaseFee() *big.Int {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if len(g.baseFeeQueue) == 0 {
		return big.NewInt(DefaultBaseFee)
	}

	// Get the sum and then the average
	total := int64(0)
	for _, fee := range g.baseFeeQueue {
		total += fee
	}

	return big.NewInt(total / int64(len(g.baseFeeQueue)))
}

// GetBaseFee returns the estimated tip (priority fee).
func (g *gasCalculator) GetTip() *big.Int {
	g.lock.RLock()
	defer g.lock.RUnlock()

	if len(g.tipQueue) == 0 {
		return big.NewInt(DefaultTip)
	}

	// Get the sum and then the average
	total := int64(0)
	for _, tip := range g.tipQueue {
		total += tip
	}

	return big.NewInt(total / int64(len(g.tipQueue)))
}

// GetGasPrice returns estimated gas price.
func (g *gasCalculator) GetGasPrice() *big.Int {
	g.lock.Lock()
	lastUpdate := g.lastUpdateGasPrice
	g.lock.Unlock()

	if time.Now().After(lastUpdate.Add(g.gasPriceUpdateInterval)) {
		// Update the gas price
		g.updateGasPrice()
	}

	g.lock.Lock()
	defer g.lock.Unlock()

	return big.NewInt(g.gasPrice)
}

func (g *gasCalculator) updateGasPrice() {
	ctx, cancel := context.WithTimeout(context.Background(), RpcTimeOut)
	gasPrice, err := g.client.SuggestGasPrice(ctx)
	cancel()

	if err != nil {
		log.Errorf("Failed to get gas price for chain %s", g.cfg.Chain)
	} else {
		g.lock.Lock()
		g.gasPrice = gasPrice.Int64()
		g.lastUpdateGasPrice = time.Now()
		g.lock.Unlock()
	}
}

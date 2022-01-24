package client

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

const (
	RETRY_TIME = 10 * time.Second
)

// A client that connects to Sisu server
type Client interface {
	TryDial()
	GetVersion() (string, error)
	BroadcastTxs(txs *types.Txs) error
	PostDeploymentResult(result *types.DispatchedTxResult) error
	UpdateGasPrice(req *types.GasPriceRequest) error
}

var (
	ErrSisuServerNotConnected = errors.New("sisu server is not connected")
)

type DefaultClient struct {
	client    *rpc.Client
	url       string
	connected bool
}

func NewClient(url string) Client {
	return &DefaultClient{
		url: url,
	}
}

func (c *DefaultClient) TryDial() {
	log.Info("Trying to dial Sisu server")

	for {
		log.Info("Dialing...", c.url)
		var err error
		c.client, err = rpc.DialContext(context.Background(), c.url)
		if err != nil {
			log.Error("Cannot connect to Sisu server err = ", err)
			time.Sleep(RETRY_TIME)
			continue
		}

		_, err = c.GetVersion()
		if err != nil {
			log.Error("Cannot get Sisu version err = ", err)
			time.Sleep(RETRY_TIME)
			continue
		}

		c.connected = true
		break
	}

	log.Info("Sisu server is connected")
}

func (c *DefaultClient) GetVersion() (string, error) {
	var version string
	err := c.client.CallContext(context.Background(), &version, "tss_version")
	return version, err
}

// TODO: Handle the case when broadcasting fails. In that case, we need to save the first Tx
// that we need to send to Sisu.
func (c *DefaultClient) BroadcastTxs(txs *types.Txs) error {
	log.Verbose("Broadcasting to Sisu server...")

	var result string
	err := c.client.CallContext(context.Background(), &result, "tss_postObservedTxs", txs)
	if err != nil {
		log.Error("Cannot broadcast tx to Sisu, err = ", err)
		return err
	}
	log.Verbose("Done broadcasting!")

	return nil
}

func (c *DefaultClient) PostDeploymentResult(result *types.DispatchedTxResult) error {
	log.Verbose("Sending Tx Deployment result back to Sisu...")

	var r string
	err := c.client.CallContext(context.Background(), &r, "tss_postDeploymentResult", result)
	if err != nil {
		log.Error("Cannot post tx deployment to sisu", "tx hash =", result.TxHash, "err = ", err)
		return err
	}

	return nil
}

func (c *DefaultClient) UpdateGasPrice(request *types.GasPriceRequest) error {
	log.Debug("Posting gas price back to Sisu...")

	var r string
	err := c.client.CallContext(context.Background(), &r, "tss_updateGasPrice", request)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

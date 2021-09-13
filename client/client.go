package client

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

const (
	RETRY_TIME = 10 * time.Second
)

var (
	SISU_SERVER_NOT_CONNECTED = errors.New("Sisu server is not connected")
)

// A client that connects to Sisu server
type Client interface {
	TryDial()
	BroadcastTxs(txs *types.Txs) error
}

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
	utils.LogInfo("Trying to dial Sisu server")

	for {
		utils.LogInfo("Dialing...", c.url)
		var err error
		c.client, err = rpc.DialContext(context.Background(), c.url)
		if err != nil {
			utils.LogError("Cannot connect to Sisu server err = ", err)
			time.Sleep(RETRY_TIME)
			continue
		}

		_, err = c.GetVersion()
		if err != nil {
			utils.LogError("Cannot get Sisu version err = ", err)
			time.Sleep(RETRY_TIME)
			continue
		}

		c.connected = true
		break
	}

	utils.LogInfo("Sisu server is connected")
}

func (c *DefaultClient) GetVersion() (string, error) {
	var version string
	err := c.client.CallContext(context.Background(), &version, "tss_version")
	return version, err
}

// TODO: Handle the case when broadcasting fails. In that case, we need to save the first Tx
// that we need to send to Sisu.
func (c *DefaultClient) BroadcastTxs(txs *types.Txs) error {
	utils.LogVerbose("Broadcasting to Sisu server...")

	var result string
	err := c.client.CallContext(context.Background(), &result, "tss_postObservedTxs", txs)
	if err != nil {
		utils.LogError("Cannot broadcast tx to Sisu, err = ", err)
		return err
	}

	return nil
}

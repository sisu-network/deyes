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
type Client struct {
	client    *rpc.Client
	url       string
	connected bool
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) TryDial() {
	utils.LogInfo("Trying to dial Sisu server")

	for {
		utils.LogInfo("Dialing...", c.url)
		var err error
		c.client, err = rpc.DialContext(context.Background(), c.url)
		if err == nil {
			c.connected = true
			break
		}
		time.Sleep(RETRY_TIME)
	}

	utils.LogInfo("Sisu server is connected")
}

func (c *Client) BroadcastTxs(txs *types.Txs) {
	utils.LogVerbose("Broadcasting to Sisu server...")
}

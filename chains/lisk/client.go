package lisk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/deyes/config"
)

type APIErr struct {
	message string
}

func NewApiErr(message string) error {
	return &APIErr{message: message}
}

func (e *APIErr) Error() string {
	return fmt.Sprintf(e.message)
}

// LiskClient  A wrapper around lisk.client so that we can mock in watcher tests.
type LiskClient interface {
	BlockNumber() (uint64, error)
	BlockByHeight(height uint64) (*types.Block, error)
	TransactionByBlock(block string) ([]*types.Transaction, error)
}

type defaultLiskClient struct {
	chain string
	rpc   string
}

func NewLiskClient(cfg config.Chain) LiskClient {
	c := &defaultLiskClient{
		chain: cfg.Chain,
		rpc:   cfg.Rpcs[0],
	}
	return c
}

func (c *defaultLiskClient) execute(endpoint string, params map[string]string) ([]byte, error) {
	keys := reflect.ValueOf(params).MapKeys()
	req, err := http.NewRequest("GET", c.rpc+endpoint, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for _, key := range keys {
		q.Add(key.Interface().(string), params[key.Interface().(string)])
	}

	req.URL.RawQuery = q.Encode()
	response, err := http.Get(req.URL.String())
	if response == nil {
		return nil, NewApiErr("cannot fetch data " + endpoint)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseData, err
}

func (c *defaultLiskClient) BlockNumber() (uint64, error) {
	params := map[string]string{
		"limit": "1",
		"sort":  "height:desc",
	}
	response, err := c.execute("/blocks", params)
	if err != nil {
		return 0, err
	}

	var responseObject types.ResponseBlock
	err = json.Unmarshal(response, &responseObject)
	if err != nil {
		return 0, err
	}
	blocks := responseObject.Data
	latestBlock := blocks[0]
	return latestBlock.Height, nil
}

func (c *defaultLiskClient) BlockByHeight(height uint64) (*types.Block, error) {
	params := map[string]string{
		"limit":  "1",
		"height": strconv.FormatUint(uint64(height), 10),
	}
	response, err := c.execute("/blocks", params)
	var responseObject types.ResponseBlock
	err = json.Unmarshal(response, &responseObject)
	if err != nil {
		return nil, err
	}

	blocks := responseObject.Data
	if len(blocks) == 0 {
		return nil, NewApiErr("lisk block is not found")
	}
	latestBlock := blocks[0]

	return &latestBlock, err
}

func (c *defaultLiskClient) TransactionByBlock(block string) ([]*types.Transaction, error) {
	params := map[string]string{
		"blockId": block,
	}
	response, err := c.execute("/transactions", params)
	if err != nil {
		return nil, err
	}

	var responseObject types.ResponseTransaction
	err = json.Unmarshal(response, &responseObject)
	if err != nil {
		return nil, err
	}

	return responseObject.Data, nil
}

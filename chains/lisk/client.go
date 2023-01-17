package lisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/sisu-network/lib/log"

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

// Client  A wrapper around lisk.client so that we can mock in watcher tests.
type Client interface {
	BlockNumber() (uint64, error)
	BlockByHeight(height uint64) (*types.Block, error)
	TransactionByBlock(block string) ([]*types.Transaction, error)
	GetAccount(address string) (*types.Account, error)
	CreateTransaction(txHash string) (string, error)
}

type defaultClient struct {
	chain string
	rpc   string
}

func NewLiskClient(cfg config.Chain) Client {
	c := &defaultClient{
		chain: cfg.Chain,
		rpc:   cfg.Rpcs[0],
	}
	return c
}

func (c *defaultClient) get(endpoint string, params map[string]string) ([]byte, error) {
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

func (c *defaultClient) post(endpoint string, body map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.rpc+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *defaultClient) BlockNumber() (uint64, error) {
	params := map[string]string{
		"limit": "1",
		"sort":  "height:desc",
	}
	response, err := c.get("/blocks", params)
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

func (c *defaultClient) CreateTransaction(txHash string) (string, error) {
	params := map[string]string{
		"transaction": txHash,
	}
	response, err := c.post("/transactions", params)
	if err != nil {
		return "", err
	}

	fmt.Println("Lisk response = ", string(response))

	var responseObject types.TransactionResponse
	err = json.Unmarshal(response, &responseObject)
	if err != nil {
		return "", err
	}
	message := responseObject.TransactionId

	return message, nil
}

func (c *defaultClient) BlockByHeight(height uint64) (*types.Block, error) {
	params := map[string]string{
		"limit":  "1",
		"height": strconv.FormatUint(uint64(height), 10),
	}
	response, err := c.get("/blocks", params)
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

	return latestBlock, err
}

func (c *defaultClient) TransactionByBlock(block string) ([]*types.Transaction, error) {
	params := map[string]string{
		"blockId": block,
	}
	response, err := c.get("/transactions", params)
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

func (c *defaultClient) GetAccount(address string) (*types.Account, error) {
	params := map[string]string{
		"address": address,
	}
	response, err := c.get("/accounts", params)
	if err != nil {
		return nil, err
	}

	var responseObject types.ResponseAccount
	err = json.Unmarshal(response, &responseObject)
	if err != nil {
		log.Errorf("GetAccount: Failed to marshal response, err = %s", err)
		return nil, err
	}

	accounts := responseObject.Data
	if len(accounts) == 0 {
		return nil, NewApiErr("lisk account block is not found")
	}
	validAccount := accounts[0]

	return validAccount, nil
}

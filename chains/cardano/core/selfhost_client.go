package core

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/sisu-network/deyes/types"

	_ "github.com/lib/pq"
)

var _ CardanoClient = (*SelfHostClient)(nil)

type PostgresConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	DbName   string `json:"schema,omitempty"`
}

type SelfHostClient struct {
	Cfg       PostgresConfig
	SubmitURL string

	Connector *SyncDB
}

func NewSelfHostClient(cfg PostgresConfig, submitURL string) *SelfHostClient {
	c := &SelfHostClient{
		Cfg:       cfg,
		SubmitURL: submitURL,
	}

	c.connectDB()
	return c
}

func ConnectDB(cfg PostgresConfig) (*sql.DB, error) {
	dbSrc := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)
	db, err := sql.Open("postgres", dbSrc)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *SelfHostClient) connectDB() {
	db, err := ConnectDB(s.Cfg)
	if err != nil {
		panic(err)
	}

	s.Connector = NewSyncDBConnector(db)
}

func (s *SelfHostClient) IsHealthy() bool {
	//TODO implement me
	panic("implement me")
}

func (s *SelfHostClient) LatestBlock() *blockfrost.Block {
	//TODO implement me
	panic("implement me")
}

func (s *SelfHostClient) GetBlock(hashOrNumber string) (*blockfrost.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SelfHostClient) BlockHeight() (int, error) {
	return s.Connector.BlockHeight()
}

func (s *SelfHostClient) NewTxs(fromHeight int, interestedAddrs map[string]bool) ([]*types.CardanoTxInItem, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SelfHostClient) SubmitTx(tx *cardano.Tx) (*cardano.Hash32, error) {
	url := fmt.Sprintf("%s/api/tx/submit", s.SubmitURL)
	txBytes := tx.Bytes()

	req, err := http.NewRequest("POST", url, bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/cbor")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

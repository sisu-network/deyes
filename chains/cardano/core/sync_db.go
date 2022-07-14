package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/config"

	_ "github.com/lib/pq"
)

var _ Provider = (*SyncDB)(nil)

type SyncDB struct {
	DB *sql.DB
}

func NewSyncDBConnector(db *sql.DB) *SyncDB {
	return &SyncDB{DB: db}
}

func ConnectDB(cfg config.SyncDbConfig) (*sql.DB, error) {
	dbSrc := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName)
	db, err := sql.Open("postgres", dbSrc)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *SyncDB) Health(ctx context.Context) (blockfrost.Health, error) {
	return blockfrost.Health{IsHealthy: true}, nil
}

func (s *SyncDB) BlockLatest(ctx context.Context) (blockfrost.Block, error) {
	rows, err := s.DB.Query("SELECT block_no FROM block ORDER BY id DESC LIMIT 1")
	if err != nil {
		return blockfrost.Block{}, err
	}

	defer rows.Close()
	rows.Next()
	var height sql.NullInt64
	if err := rows.Scan(&height); err != nil {
		return blockfrost.Block{}, err
	}

	return blockfrost.Block{Height: int(height.Int64)}, nil
}

func (s *SyncDB) Block(ctx context.Context, hashOrNumber string) (blockfrost.Block, error) {
	num, err := strconv.Atoi(hashOrNumber)
	if err != nil {
		return blockfrost.Block{}, err
	}

	rows, err := s.DB.Query("select id from block where block_no = $1", num)
	if err != nil {
		return blockfrost.Block{}, err
	}

	if !rows.Next() {
		err := fmt.Errorf("block %d not found", num)
		return blockfrost.Block{}, err
	}

	return blockfrost.Block{Height: num}, nil
}

func (s *SyncDB) AddressTransactions(ctx context.Context, address string, query blockfrost.APIQueryParams) ([]blockfrost.AddressTransactions, error) {
	from, err := strconv.Atoi(query.From)
	if err != nil {
		return nil, err
	}

	dbQuery := "select encode(hash, 'hex') from tx where block_id = (select id from block where block_no = $1) and id in (select tx_id from tx_out where address = $2)"
	rows, err := s.DB.Query(dbQuery, from, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txs := make([]string, 0)
	for rows.Next() {
		var rc sql.NullString
		if err := rows.Scan(&rc); err != nil {
			return nil, err
		}

		txs = append(txs, rc.String)
	}

	res := make([]blockfrost.AddressTransactions, 0, len(txs))
	for _, tx := range txs {
		res = append(res, blockfrost.AddressTransactions{TxHash: tx})
	}

	return res, nil
}

func (s *SyncDB) BlockTransactions(_ context.Context, height string) ([]blockfrost.Transaction, error) {
	h, err := strconv.Atoi(height)
	if err != nil {
		return []blockfrost.Transaction{}, err
	}

	rows, err := s.DB.Query("SELECT encode(hash, 'hex') FROM tx WHERE block_id in (SELECT id FROM block WHERE block_no = $1)", h)
	if err != nil {
		return []blockfrost.Transaction{}, err
	}

	defer rows.Close()
	txs := make([]blockfrost.Transaction, 0)
	for rows.Next() {
		var rc sql.NullString
		if err := rows.Scan(&rc); err != nil {
			return []blockfrost.Transaction{}, err
		}

		txs = append(txs, blockfrost.Transaction(rc.String))
	}

	return txs, nil
}

func (s *SyncDB) BlockHeight() (int, error) {
	row, err := s.DB.Query("SELECT block_no FROM block ORDER BY id DESC LIMIT 1")
	if err != nil {
		return 0, err
	}

	defer row.Close()
	var blockNum sql.NullInt64
	row.Next()
	if err := row.Scan(&blockNum); err != nil {
		return 0, err
	}

	return int(blockNum.Int64), nil
}

func (s *SyncDB) TransactionUTXOs(ctx context.Context, hash string) (blockfrost.TransactionUTXOs, error) {
	hash = "\\x" + hash
	res, err := s.GetTxOutIDs(ctx, hash)
	if err != nil {
		return blockfrost.TransactionUTXOs{}, err
	}

	outputs := make([]struct {
		Address string                `json:"address"`
		Amount  []blockfrost.TxAmount `json:"amount"`
	}, 0)

	for index, id := range res.Ids {
		rows, err := s.DB.Query("SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id = $1", id)
		if err != nil {
			return blockfrost.TransactionUTXOs{}, err
		}

		txAmounts := make([]blockfrost.TxAmount, 0)
		for rows.Next() {
			var quantity sql.NullString
			var maID sql.NullInt64
			if err := rows.Scan(&quantity, &maID); err != nil {
				return blockfrost.TransactionUTXOs{}, err
			}

			maRow, err := s.DB.Query("SELECT encode(policy, 'hex'), encode(name, 'hex') FROM multi_asset where id = $1", maID.Int64)
			if err != nil {
				return blockfrost.TransactionUTXOs{}, err
			}

			maRow.Next()
			var policy, maName sql.NullString
			if err := maRow.Scan(&policy, &maName); err != nil {
				return blockfrost.TransactionUTXOs{}, err
			}
			maRow.Close()

			txAmount := blockfrost.TxAmount{
				Quantity: quantity.String,
				Unit:     policy.String + maName.String,
			}
			txAmounts = append(txAmounts, txAmount)
		}
		rows.Close()

		txAmounts = append(txAmounts, blockfrost.TxAmount{
			Quantity: strconv.FormatInt(res.Values[index], 10),
			Unit:     "lovelace",
		})

		outputs = append(outputs, struct {
			Address string                `json:"address"`
			Amount  []blockfrost.TxAmount `json:"amount"`
		}{
			Address: res.Addresses[index],
			Amount:  txAmounts,
		})
	}

	return blockfrost.TransactionUTXOs{
		Hash:    hash,
		Outputs: outputs,
	}, nil
}

type GetTxOutIDsResult struct {
	Ids, Values []int64
	Addresses   []string
}

func (s *SyncDB) GetTxOutIDs(_ context.Context, hash string) (GetTxOutIDsResult, error) {
	query := `SELECT id, address, value FROM tx_out WHERE tx_id = (SELECT id FROM tx WHERE hash = '` + hash + `') ORDER BY id`
	rows, err := s.DB.Query(query)
	if err != nil {
		return GetTxOutIDsResult{}, err
	}

	defer rows.Close()
	var (
		ids, values []int64
		addrs       []string
	)
	for rows.Next() {
		var id, value sql.NullInt64
		var address sql.NullString
		if err := rows.Scan(&id, &address, &value); err != nil {
			return GetTxOutIDsResult{}, err
		}

		ids = append(ids, id.Int64)
		addrs = append(addrs, address.String)
		values = append(values, value.Int64)
	}

	return GetTxOutIDsResult{
		Ids:       ids,
		Addresses: addrs,
		Values:    values,
	}, nil
}

func (s *SyncDB) TransactionMetadata(_ context.Context, hash string) ([]blockfrost.TransactionMetadata, error) {
	hash = "\\x" + hash
	query := `SELECT key, json FROM tx_metadata WHERE tx_id = (SELECT id FROM tx WHERE hash ='` + hash + `')`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	res := make([]blockfrost.TransactionMetadata, 0)
	for rows.Next() {
		var label, metadata sql.NullString
		if err := rows.Scan(&label, &metadata); err != nil {
			return nil, err
		}

		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(metadata.String), &m); err == nil {
			res = append(res, blockfrost.TransactionMetadata{
				JsonMetadata: m,
				Label:        label.String,
			})
		} else {
			res = append(res, blockfrost.TransactionMetadata{
				JsonMetadata: metadata.String,
				Label:        label.String,
			})
		}
	}

	return res, nil
}

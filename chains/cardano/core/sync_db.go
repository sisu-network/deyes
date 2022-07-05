package core

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
)

type SyncDB struct {
	DB *sql.DB
}

func NewSyncDBConnector(db *sql.DB) *SyncDB {
	return &SyncDB{DB: db}
}

func (s *SyncDB) BlockTransactions(ctx context.Context, height int64) ([]string, error) {
	rows, err := s.DB.Query("SELECT encode(hash, 'hex') FROM tx WHERE block_id in (SELECT id FROM block WHERE block_no = $1)", height)
	if err != nil {
		return []string{}, err
	}

	defer rows.Close()
	txs := make([]string, 0)
	for rows.Next() {
		var rc sql.NullString
		if err := rows.Scan(&rc); err != nil {
			return []string{}, err
		}

		txs = append(txs, rc.String)
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
	txOutIds, err := s.GetTxOutIDs(ctx, hash)
	if err != nil {
		return blockfrost.TransactionUTXOs{}, err
	}

	arr := "("
	for i, id := range txOutIds {
		arr += strconv.Itoa(int(id))
		if i != len(txOutIds)-1 {
			arr += ","
		}
	}
	arr += ")"

	query := "SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id IN " + arr
	rows, err := s.DB.Query(query)
	if err != nil {
		return blockfrost.TransactionUTXOs{}, err
	}

	for rows.Next() {
		var quantity, maID sql.NullInt64
		if err := rows.Scan(&quantity, &maID); err != nil {
			return blockfrost.TransactionUTXOs{}, err
		}
	}

	return blockfrost.TransactionUTXOs{}, nil
}

func (s *SyncDB) GetTxOutIDs(_ context.Context, hash string) ([]int64, error) {
	query := `SELECT id FROM tx_out WHERE tx_id = (SELECT id FROM tx WHERE hash = '` + hash + `')`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var r sql.NullInt64
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}

		ids = append(ids, r.Int64)
	}

	return ids, nil
}

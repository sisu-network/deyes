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
				Unit:     maName.String + policy.String,
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

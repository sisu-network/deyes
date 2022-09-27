package cardano

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/blockfrost/blockfrost-go"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"

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

func (s *SyncDB) Health(ctx context.Context) (bool, error) {
	// TODO: Check connection with db sync.
	return true, nil
}

func (s *SyncDB) BlockLatest(ctx context.Context) (*types.CardanoBlock, error) {
	rows, err := s.DB.Query("SELECT block_no, epoch_no, slot_no FROM block ORDER BY id DESC LIMIT 1")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	rows.Next()
	var height, epoch, slot sql.NullInt64
	if err := rows.Scan(&height, &epoch, &slot); err != nil {
		return nil, err
	}

	return &types.CardanoBlock{
		Height: int(height.Int64),
		Epoch:  int(epoch.Int64),
		Slot:   int(slot.Int64),
	}, nil
}

func (s *SyncDB) Block(ctx context.Context, hashOrNumber string) (*types.CardanoBlock, error) {
	num, err := strconv.Atoi(hashOrNumber)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query("select id from block where block_no = $1", num)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		err := fmt.Errorf("block %d not found", num)
		return nil, err
	}

	return &types.CardanoBlock{Height: num}, nil
}

func (s *SyncDB) AddressTransactions(ctx context.Context, address string, query types.APIQueryParams) ([]*types.AddressTransactions, error) {
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

	res := make([]*types.AddressTransactions, 0, len(txs))
	for _, tx := range txs {
		res = append(res, &types.AddressTransactions{TxHash: tx})
	}

	return res, nil
}

func (s *SyncDB) BlockTransactions(_ context.Context, height string) ([]string, error) {
	h, err := strconv.Atoi(height)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query("SELECT encode(hash, 'hex') FROM tx WHERE block_id in (SELECT id FROM block WHERE block_no = $1)", h)
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

func (s *SyncDB) TransactionUTXOs(ctx context.Context, hash string) (*types.TransactionUTXOs, error) {
	hash = "\\x" + hash
	res, err := s.GetTxOutIDs(ctx, hash)
	if err != nil {
		return nil, err
	}

	outputs := make([]types.TransactionUTXOsOutput, 0)

	for index, id := range res.Ids {
		rows, err := s.DB.Query("SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id = $1", id)
		if err != nil {
			return nil, err
		}

		txAmounts := make([]types.TxAmount, 0)
		for rows.Next() {
			var quantity sql.NullString
			var maID sql.NullInt64
			if err := rows.Scan(&quantity, &maID); err != nil {
				return nil, err
			}

			maRow, err := s.DB.Query("SELECT encode(policy, 'hex'), encode(name, 'hex') FROM multi_asset where id = $1", maID.Int64)
			if err != nil {
				return nil, err
			}

			maRow.Next()
			var policy, maName sql.NullString
			if err := maRow.Scan(&policy, &maName); err != nil {
				return nil, err
			}
			maRow.Close()

			txAmount := types.TxAmount{
				Quantity: quantity.String,
				Unit:     policy.String + maName.String,
			}
			txAmounts = append(txAmounts, txAmount)
		}
		rows.Close()

		txAmounts = append(txAmounts, types.TxAmount{
			Quantity: strconv.FormatInt(res.Values[index], 10),
			Unit:     "lovelace",
		})

		outputs = append(outputs, struct {
			Address string           `json:"address"`
			Amount  []types.TxAmount `json:"amount"`
		}{
			Address: res.Addresses[index],
			Amount:  txAmounts,
		})
	}

	return &types.TransactionUTXOs{
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

func (s *SyncDB) TransactionMetadata(_ context.Context, hash string) ([]*types.TransactionMetadata, error) {
	hash = "\\x" + hash
	query := `SELECT key, json FROM tx_metadata WHERE tx_id = (SELECT id FROM tx WHERE hash ='` + hash + `')`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	res := make([]*types.TransactionMetadata, 0)
	for rows.Next() {
		var label, metadata sql.NullString
		if err := rows.Scan(&label, &metadata); err != nil {
			return nil, err
		}

		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(metadata.String), &m); err == nil {
			res = append(res, &types.TransactionMetadata{
				JsonMetadata: m,
				Label:        label.String,
			})
		} else {
			res = append(res, &types.TransactionMetadata{
				JsonMetadata: metadata.String,
				Label:        label.String,
			})
		}
	}

	return res, nil
}

func (s *SyncDB) LatestEpochParameters(ctx context.Context) (types.EpochParameters, error) {
	query := "SELECT min_fee_a, min_fee_b, max_block_size, max_tx_size, max_bh_size, key_deposit, pool_deposit," +
		" max_epoch, optimal_pool_count, coins_per_utxo_size FROM epoch_param ORDER BY id DESC LIMIT 1"
	rows, err := s.DB.Query(query)
	if err != nil {
		return types.EpochParameters{}, err
	}

	defer rows.Close()
	rows.Next()
	var (
		minFeeA, minFeeB, maxBlockSize, maxTxSize, maxBlockHeaderSize, maxEpoch, nopt sql.NullInt64
		keyDeposit, poolDeposit, minUTXOValue                                         sql.NullString
	)
	if err := rows.Scan(&minFeeA, &minFeeB, &maxBlockSize, &maxTxSize, &maxBlockHeaderSize, &keyDeposit,
		&poolDeposit, &maxEpoch, &nopt, &minUTXOValue); err != nil {
		return types.EpochParameters{}, err
	}

	return types.EpochParameters{
		Epoch:              int(maxEpoch.Int64),
		KeyDeposit:         keyDeposit.String,
		MaxBlockHeaderSize: int(maxBlockHeaderSize.Int64),
		MaxBlockSize:       int(maxBlockSize.Int64),
		MaxTxSize:          int(maxTxSize.Int64),
		MinFeeA:            int(minFeeA.Int64),
		MinFeeB:            int(minFeeB.Int64),
		MinUtxo:            minUTXOValue.String,
		NOpt:               int(nopt.Int64),
		PoolDeposit:        poolDeposit.String,
	}, nil
}

type TxOut struct {
	ID      int
	TxID    int
	Index   int
	Address string
	Value   string
}

type TxIn struct {
	ID         int
	TxInID     int
	TxOutID    int
	TxOutIndex int
}

type Tx struct {
	ID      int
	Hash    string
	BlockID int
}

type TxInfo struct {
	TxHash  string
	BlockID int
}

// AddressUTXOs queries address's utxos at specific block height
func (s *SyncDB) AddressUTXOs(ctx context.Context, address string, query types.APIQueryParams) ([]types.AddressUTXO, error) {
	var maxBlock int
	if len(query.To) == 0 {
		maxBlock = math.MaxInt32
	} else {
		to, err := strconv.ParseUint(query.To, 10, 64)
		if err != nil {
			return nil, err
		}

		maxBlock = int(to)
		// check overflow here because postgresql only support int32 type
		if to > math.MaxInt32 {
			maxBlock = math.MaxInt32
		}
	}

	txIDs, txInfoMap, err := s.getAddressTxsUntilMaxBlock(address, maxBlock)
	if err != nil {
		return nil, err
	}

	if len(txIDs) == 0 {
		return nil, nil
	}

	txOuts, err := s.getTxOutByTxID(address, txIDs)
	if err != nil {
		return nil, err
	}

	txIns, err := s.getTxInByTxID(txIDs)
	if err != nil {
		return nil, err
	}

	unusedTxOuts := make([]TxOut, 0)
	for _, txOut := range txOuts {
		used := false
		for _, txIn := range txIns {
			if txOut.TxID == txIn.TxOutID && txOut.Index == txIn.TxOutIndex {
				used = true
				break
			}
		}

		if !used {
			unusedTxOuts = append(unusedTxOuts, txOut)
		}
	}

	res := make([]types.AddressUTXO, 0)

	for _, txOut := range unusedTxOuts {
		addressAmounts, err := s.getAddressAmountFromTxOut(txOut)
		if err != nil {
			return nil, err
		}

		block, err := s.GetBlockByID(ctx, txInfoMap[txOut.TxID].BlockID)
		if err != nil {
			return nil, err
		}

		res = append(res, types.AddressUTXO{
			TxHash:      txInfoMap[txOut.TxID].TxHash,
			OutputIndex: txOut.Index,
			Amount:      addressAmounts,
			Block:       block.Hash,
		})
	}

	return res, nil
}

func (s *SyncDB) getAddressTxsUntilMaxBlock(address string, maxBlock int) ([]int64, map[int]TxInfo, error) {
	txQuery := "select id, encode(hash, 'hex'), block_id from tx where block_id in (select id from block where block_no <= $1) and id in (select tx_id from tx_out where address = $2)"
	rows, err := s.DB.Query(txQuery, maxBlock, address)
	if err != nil {
		return nil, nil, err
	}

	txIDs := make([]int64, 0)
	txInfoMap := make(map[int]TxInfo)

	for rows.Next() {
		var id, blockID sql.NullInt64
		var hash sql.NullString
		if err := rows.Scan(&id, &hash, &blockID); err != nil {
			return nil, nil, err
		}

		txIDs = append(txIDs, id.Int64)
		txInfoMap[int(id.Int64)] = TxInfo{
			TxHash:  hash.String,
			BlockID: int(blockID.Int64),
		}
	}
	rows.Close()

	return txIDs, txInfoMap, nil
}

func (s *SyncDB) getAddressAmountFromTxOut(txOut TxOut) ([]types.AddressAmount, error) {
	rows, err := s.DB.Query("SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id = $1", txOut.ID)
	if err != nil {
		return nil, err
	}

	addressAmounts := make([]types.AddressAmount, 0)
	for rows.Next() {
		var quantity sql.NullString
		var maID sql.NullInt64
		if err := rows.Scan(&quantity, &maID); err != nil {
			return nil, err
		}

		maRow, err := s.DB.Query("SELECT encode(policy, 'hex'), encode(name, 'hex') FROM multi_asset where id = $1", maID.Int64)
		if err != nil {
			return nil, err
		}

		maRow.Next()
		var policy, maName sql.NullString
		if err := maRow.Scan(&policy, &maName); err != nil {
			return nil, err
		}
		maRow.Close()

		amount := types.AddressAmount{
			Quantity: quantity.String,
			Unit:     policy.String + maName.String,
		}
		addressAmounts = append(addressAmounts, amount)
	}
	defer rows.Close()

	addressAmounts = append(addressAmounts, types.AddressAmount{
		Unit:     "lovelace",
		Quantity: txOut.Value,
	})

	return addressAmounts, nil
}

func (s *SyncDB) getTxOutByTxID(address string, txIDs []int64) ([]TxOut, error) {
	str := buildQueryFromIntArray(txIDs)
	txOutQuery := "select id, tx_id, index, address, value from tx_out where tx_id in " + str + " and address = $1"
	rows, err := s.DB.Query(txOutQuery, address)
	if err != nil {
		return nil, err
	}

	txOuts := make([]TxOut, 0)
	for rows.Next() {
		var (
			id, txId, index sql.NullInt64
			address, value  sql.NullString
		)

		if err := rows.Scan(&id, &txId, &index, &address, &value); err != nil {
			return nil, err
		}

		txOuts = append(txOuts, TxOut{
			ID:      int(id.Int64),
			TxID:    int(txId.Int64),
			Index:   int(index.Int64),
			Address: address.String,
			Value:   value.String,
		})
	}
	defer rows.Close()

	return txOuts, nil
}

func (s *SyncDB) getTxInByTxID(txIDs []int64) ([]TxIn, error) {
	str := buildQueryFromIntArray(txIDs)
	txInQuery := "select id, tx_in_id, tx_out_id, tx_out_index from tx_in where tx_out_id in " + str
	rows, err := s.DB.Query(txInQuery)
	if err != nil {
		return nil, err
	}

	txIns := make([]TxIn, 0)
	for rows.Next() {
		var id, txInID, txOutID, txOutIndex sql.NullInt64
		if err := rows.Scan(&id, &txInID, &txOutID, &txOutIndex); err != nil {
			return nil, err
		}

		txIns = append(txIns, TxIn{
			ID:         int(id.Int64),
			TxInID:     int(txInID.Int64),
			TxOutID:    int(txOutID.Int64),
			TxOutIndex: int(txOutIndex.Int64),
		})
	}
	defer rows.Close()

	return txIns, nil
}

func (s *SyncDB) GetBlockByID(_ context.Context, id int) (blockfrost.Block, error) {
	query := "select encode(hash, 'hex'), block_no, slot_no, epoch_no from block where id = $1 limit 1"
	row, err := s.DB.Query(query, id)
	if err != nil {
		return blockfrost.Block{}, err
	}

	row.Next()
	var (
		blockNumber, slot, epoch sql.NullInt64
		blockHash                sql.NullString
	)

	if err := row.Scan(&blockHash, &blockNumber, &slot, &epoch); err != nil {
		return blockfrost.Block{}, err
	}

	return blockfrost.Block{
		Height: int(blockNumber.Int64),
		Hash:   blockHash.String,
		Slot:   int(slot.Int64),
		Epoch:  int(epoch.Int64),
	}, nil
}

func buildQueryFromIntArray(arr []int64) string {
	strArr := make([]string, 0, len(arr))
	for _, element := range arr {
		strArr = append(strArr, strconv.Itoa(int(element)))
	}

	res := strings.Join(strArr, ",")
	res = "(" + res + ")"
	return res
}

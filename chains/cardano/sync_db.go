package cardano

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	"github.com/ethereum/go-ethereum/common/math"
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
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

func (s *SyncDB) Health(ctx context.Context) (bool, error) {
	// TODO: Check connection with db sync.
	return true, nil
}

func (s *SyncDB) BlockLatest(ctx context.Context) (*providertypes.Block, error) {
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

	return &providertypes.Block{
		Height: int(height.Int64),
		Epoch:  int(epoch.Int64),
		Slot:   int(slot.Int64),
	}, nil
}

func (s *SyncDB) Block(ctx context.Context, number string) (*providertypes.Block, error) {
	num, err := strconv.Atoi(number)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query("select encode(hash, 'hex'), block_no, slot_no, epoch_no from block where block_no = $1", num)
	if err != nil {
		return nil, err
	}

	rows.Next()
	var (
		blockNumber, slot, epoch sql.NullInt64
		hash                     sql.NullString
	)

	if err := rows.Scan(&hash, &blockNumber, &slot, &epoch); err != nil {
		return nil, err
	}

	return &providertypes.Block{
		Height: int(blockNumber.Int64),
		Hash:   hash.String,
		Slot:   int(slot.Int64),
		Epoch:  int(epoch.Int64),
	}, nil
}

func (s *SyncDB) AddressTransactions(ctx context.Context, address string, query providertypes.APIQueryParams) ([]*providertypes.AddressTransactions, error) {
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

	res := make([]*providertypes.AddressTransactions, 0, len(txs))
	for _, tx := range txs {
		res = append(res, &providertypes.AddressTransactions{TxHash: tx})
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

func (s *SyncDB) TransactionUTXOs(ctx context.Context, hash string) (*providertypes.TransactionUTXOs, error) {
	hash = "\\x" + hash
	res, err := s.GetTxOutIDs(ctx, hash)
	if err != nil {
		return nil, err
	}

	outputs := make([]providertypes.TransactionUTXOsOutput, 0)

	for index, id := range res.Ids {
		rows, err := s.DB.Query("SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id = $1", id)
		if err != nil {
			return nil, err
		}

		txAmounts := make([]providertypes.TxAmount, 0)
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

			txAmount := providertypes.TxAmount{
				Quantity: quantity.String,
				Unit:     policy.String + maName.String,
			}
			txAmounts = append(txAmounts, txAmount)
		}
		rows.Close()

		txAmounts = append(txAmounts, providertypes.TxAmount{
			Quantity: strconv.FormatInt(res.Values[index], 10),
			Unit:     "lovelace",
		})

		outputs = append(outputs, struct {
			Address string                   `json:"address"`
			Amount  []providertypes.TxAmount `json:"amount"`
		}{
			Address: res.Addresses[index],
			Amount:  txAmounts,
		})
	}

	return &providertypes.TransactionUTXOs{
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

func (s *SyncDB) TransactionMetadata(_ context.Context, hash string) ([]*providertypes.TransactionMetadata, error) {
	hash = "\\x" + hash
	query := `SELECT key, json FROM tx_metadata WHERE tx_id = (SELECT id FROM tx WHERE hash ='` + hash + `')`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	res := make([]*providertypes.TransactionMetadata, 0)
	for rows.Next() {
		var label, metadata sql.NullString
		if err := rows.Scan(&label, &metadata); err != nil {
			return nil, err
		}

		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(metadata.String), &m); err == nil {
			res = append(res, &providertypes.TransactionMetadata{
				JsonMetadata: m,
				Label:        label.String,
			})
		} else {
			res = append(res, &providertypes.TransactionMetadata{
				JsonMetadata: metadata.String,
				Label:        label.String,
			})
		}
	}

	return res, nil
}

func (s *SyncDB) LatestEpochParameters(ctx context.Context) (*cardano.ProtocolParams, error) {
	query := "SELECT min_fee_a, min_fee_b, max_block_size, max_tx_size, max_bh_size, key_deposit, pool_deposit," +
		" max_epoch, optimal_pool_count, coins_per_utxo_size FROM epoch_param ORDER BY id DESC LIMIT 1"
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	rows.Next()
	var (
		minFeeA, minFeeB, maxBlockSize, maxTxSize, maxBlockHeaderSize, maxEpoch, nopt sql.NullInt64
		keyDeposit, poolDeposit, minUTXOValue                                         sql.NullString
	)
	if err := rows.Scan(&minFeeA, &minFeeB, &maxBlockSize, &maxTxSize, &maxBlockHeaderSize, &keyDeposit,
		&poolDeposit, &maxEpoch, &nopt, &minUTXOValue); err != nil {
		return nil, err
	}

	keyDepositInt, err := strconv.ParseUint(keyDeposit.String, 10, 64)
	if err != nil {
		return nil, err
	}

	minUTXO, err := strconv.ParseUint(minUTXOValue.String, 10, 64)
	if err != nil {
		return nil, err
	}

	poolDepositInt, err := strconv.ParseUint(poolDeposit.String, 10, 64)
	if err != nil {
		return nil, err
	}

	return &cardano.ProtocolParams{
		MaxEpoch:           uint(maxEpoch.Int64),
		KeyDeposit:         cardano.Coin(keyDepositInt),
		MaxBlockHeaderSize: uint(maxBlockHeaderSize.Int64),
		MaxBlockBodySize:   uint(maxBlockSize.Int64),
		MaxTxSize:          uint(maxTxSize.Int64),
		MinFeeA:            cardano.Coin(minFeeA.Int64),
		MinFeeB:            cardano.Coin(minFeeB.Int64),
		CoinsPerUTXOWord:   cardano.Coin(minUTXO),
		NOpt:               uint(nopt.Int64),
		PoolDeposit:        cardano.Coin(poolDepositInt),
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
func (s *SyncDB) AddressUTXOs(ctx context.Context, address string, query providertypes.APIQueryParams) ([]cardano.UTxO, error) {
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

	res := make([]cardano.UTxO, 0)
	spender, err := cardano.NewAddress(address)
	if err != nil {
		return nil, err
	}

	for _, txOut := range unusedTxOuts {
		addressAmounts, err := s.getAddressAmountFromTxOut(txOut)
		if err != nil {
			return nil, err
		}

		txHash, err := cardano.NewHash32(txInfoMap[txOut.TxID].TxHash)
		if err != nil {
			return nil, err
		}

		res = append(res, cardano.UTxO{
			TxHash:  txHash,
			Index:   uint64(txOut.Index),
			Amount:  addressAmounts,
			Spender: spender,
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

func (s *SyncDB) getAddressAmountFromTxOut(txOut TxOut) (*cardano.Value, error) {
	rows, err := s.DB.Query("SELECT quantity, ident FROM ma_tx_out WHERE tx_out_id = $1", txOut.ID)
	if err != nil {
		return nil, err
	}

	coinValue, err := strconv.Atoi(txOut.Value)
	if err != nil {
		return nil, err
	}

	addressAmounts := cardano.NewValue(cardano.Coin(coinValue))

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

		bz, err := hex.DecodeString(policy.String)
		if err != nil {
			return nil, err
		}
		policyID := cardano.NewPolicyIDFromHash(cardano.Hash28(bz))
		assetName, err := hex.DecodeString(maName.String)
		if err != nil {
			return nil, err
		}

		assetValue, err := strconv.Atoi(quantity.String)
		if err != nil {
			return nil, err
		}

		currentAssets := addressAmounts.MultiAsset.Get(policyID)
		if currentAssets != nil {
			currentAssets.Set(
				cardano.NewAssetName(string(assetName)),
				cardano.BigNum(assetValue),
			)
		} else {
			addressAmounts.MultiAsset.Set(
				policyID,
				cardano.NewAssets().
					Set(
						cardano.NewAssetName(string(assetName)),
						cardano.BigNum(assetValue),
					),
			)
		}
	}

	defer rows.Close()

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

func (s *SyncDB) Tip(blockHeight uint64) (*cardano.NodeTip, error) {
	latestBlock, err := s.BlockLatest(context.Background())
	if err != nil {
		return nil, err
	}

	if blockHeight > uint64(latestBlock.Height) {
		blockHeight = uint64(latestBlock.Height)
	}

	block, err := s.Block(context.Background(), fmt.Sprintf("%d", blockHeight))
	if err != nil {
		return nil, err
	}

	return &cardano.NodeTip{
		Block: uint64(block.Height),
		Epoch: uint64(block.Epoch),
		Slot:  uint64(block.Slot),
	}, nil
}

// buildQueryFromIntArray is a function that creates string arguments for an array of integers. It
// is safe from SQL injection.
func buildQueryFromIntArray(arr []int64) string {
	strArr := make([]string, 0, len(arr))
	for _, element := range arr {
		strArr = append(strArr, strconv.Itoa(int(element)))
	}

	res := strings.Join(strArr, ",")
	res = "(" + res + ")"
	return res
}

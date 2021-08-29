package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

// A struct for saving txs into database.
type saveTxsRequest struct {
	chain       string
	blockHeight int64
	txs         *types.Txs
}

type Database struct {
	db       *sql.DB
	saveTxCh chan *saveTxsRequest
}

type dbLogger struct {
}

func (loggger *dbLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (loggger *dbLogger) Verbose() bool {
	return true
}

func NewDb() *Database {
	return &Database{
		saveTxCh: make(chan *saveTxsRequest),
	}
}

func (d *Database) Connect() error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		return fmt.Errorf("DB host cannot be empty")
	}

	portString := os.Getenv("DB_PORT")
	_, err := strconv.Atoi(portString)
	if err != nil {
		return err
	}

	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	schema := os.Getenv("DB_SCHEMA")

	// Connect to the db
	database, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, portString))
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE DATABASE IF NOT EXISTS " + schema)
	if err != nil {
		return err
	}
	database.Close()

	database, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, portString, schema))
	if err != nil {
		return err
	}

	d.db = database
	utils.LogInfo("Db is connected successfully")
	return nil
}

func (d *Database) DoMigration() error {
	driver, err := mysql.WithInstance(d.db, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/",
		"mysql",
		driver,
	)

	if err != nil {
		return err
	}

	m.Log = &dbLogger{}
	m.Up()

	return nil
}

func (d *Database) Init() error {
	err := d.Connect()
	if err != nil {
		utils.LogError("Failed to connect to DB. Err =", err)
		return err
	}

	err = d.DoMigration()
	if err != nil {
		return err
	}

	go d.listen()

	return nil
}

// Listen to request to save into datbase.
func (d *Database) listen() {
	for {
		select {
		case req := <-d.saveTxCh:
			err := d.doSave(req)
			if err != nil {
				utils.LogError("Cannot save into db, err = ", err)
			}
		}
	}
}

func (d *Database) doSave(req *saveTxsRequest) error {
	chain := req.chain
	txs := req.txs
	blockHeight := req.blockHeight

	for _, tx := range txs.Arr {
		hash := tx.Hash
		if len(hash) > 256 {
			hash = hash[:256]
		}

		_, err := d.db.Exec("INSERT IGNORE INTO transactions (chain, tx_hash, block_height, tx_bytes) VALUES (?, ?, ?, ?)",
			chain, hash, blockHeight, tx.Serialized)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	d.saveTxCh <- &saveTxsRequest{
		chain:       chain,
		blockHeight: blockHeight,
		txs:         txs,
	}
}

func (d *Database) LoadBlockHeight(chain string) (int64, error) {
	rows, err := d.db.Query("SELECT block_height FROM latest_block_height WHERE chain=?", chain)
	if err != nil {
		return 0, err
	}

	if !rows.Next() {
		return 0, nil
	}

	var blockHeight int64
	switch err := rows.Scan(&blockHeight); err {
	case nil:
		return blockHeight, nil
	default:
		return 0, err
	}
}

type DefaultDatabase struct {
	blockHeights map[string]int64
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{blockHeights: make(map[string]int64)}
}

func (d *DefaultDatabase) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	d.blockHeights[chain] = blockHeight
}

func (d *DefaultDatabase) LoadBlockHeight(chain string) (int64, error) {
	bh, ok := d.blockHeights[chain]
	if !ok {
		return 0, errors.New("chain not found")
	}

	return bh, nil
}

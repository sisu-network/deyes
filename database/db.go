package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/deyes/utils"
)

type Database interface {
	Init() error
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)
	LoadBlockHeight(chain string) (int64, error)

	// Watch address
	SaveWatchAddress(chain, address string)
	LoadWatchAddresses(chain string) []string
}

// A struct for saving txs into database.
type saveTxsRequest struct {
	chain       string
	blockHeight int64
	txs         *types.Txs
}

type DefaultDatabase struct {
	cfg      *config.Deyes
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

func NewDb(cfg *config.Deyes) Database {
	return &DefaultDatabase{
		cfg:      cfg,
		saveTxCh: make(chan *saveTxsRequest),
	}
}

func (d *DefaultDatabase) Connect() error {
	host := d.cfg.DbHost
	if host == "" {
		return fmt.Errorf("DB host cannot be empty")
	}

	port := d.cfg.DbPort

	username := d.cfg.DbUsername
	password := d.cfg.DbPassword
	schema := d.cfg.DbSchema

	// Connect to the db
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/", username, password, host, port)
	database, err := sql.Open("mysql", url)
	if err != nil {
		return err
	}
	_, err = database.Exec("CREATE DATABASE IF NOT EXISTS " + schema)
	if err != nil {
		return err
	}
	database.Close()

	database, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, schema))
	if err != nil {
		return err
	}

	d.db = database
	utils.LogInfo("Db is connected successfully")
	return nil
}

func (d *DefaultDatabase) DoMigration() error {
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

func (d *DefaultDatabase) Init() error {
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
func (d *DefaultDatabase) listen() {
	for req := range d.saveTxCh {
		err := d.doSave(req)
		if err != nil {
			utils.LogError("Cannot save into db, err = ", err)
		}
	}
}

func (d *DefaultDatabase) doSave(req *saveTxsRequest) error {
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

func (d *DefaultDatabase) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	d.saveTxCh <- &saveTxsRequest{
		chain:       chain,
		blockHeight: blockHeight,
		txs:         txs,
	}
}

func (d *DefaultDatabase) LoadBlockHeight(chain string) (int64, error) {
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

func (d *DefaultDatabase) SaveWatchAddress(chain, address string) {
	_, err := d.db.Exec("INSERT IGNORE INTO watch_address (chain, address) VALUES (?, ?)", chain, address)
	if err != nil {
		utils.LogError(fmt.Sprintf("cannot insert watch address with chain %s and address %s.", chain, address), "Err =", err)
	}
}

func (d *DefaultDatabase) LoadWatchAddresses(chain string) []string {
	addrs := make([]string, 0)
	rows, err := d.db.Query("SELECT address FROM watch_address WHERE chain=?", chain)
	if err != nil {
		utils.LogError("Failed to load watch address for chain", chain, ". Error = ", err)
		return addrs
	}

	for rows.Next() {
		var addr string
		err := rows.Scan(&addr)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

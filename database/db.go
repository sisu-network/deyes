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
	"github.com/sisu-network/lib/log"
)

// go:generate mockgen -source database/db.go -destination=tests/mock/network/http.go -package=mock
type Database interface {
	Init() error
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)

	// Watch address
	SaveWatchAddress(chain, address string)
	LoadWatchAddresses(chain string) []string

	// Token price
	SaveTokenPrices(tokenPrices types.TokenPrices)
	LoadPrices() types.TokenPrices
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
	log.Info("Db is connected successfully")
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
		log.Error("Failed to connect to DB. Err =", err)
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
	for {
		select {
		case req := <-d.saveTxCh:
			err := d.doSave(req)
			if err != nil {
				log.Error("Cannot save into db, err = ", err)
			}
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

func (d *DefaultDatabase) SaveWatchAddress(chain, address string) {
	_, err := d.db.Exec("INSERT IGNORE INTO watch_address (chain, address) VALUES (?, ?)", chain, address)
	if err != nil {
		log.Error(fmt.Sprintf("cannot insert watch address with chain %s and address %s.", chain, address), "Err =", err)
	}
}

func (d *DefaultDatabase) LoadWatchAddresses(chain string) []string {
	addrs := make([]string, 0)
	rows, err := d.db.Query("SELECT address FROM watch_address WHERE chain=?", chain)
	if err != nil {
		log.Error("Failed to load watch address for chain", chain, ". Error = ", err)
		return addrs
	}

	defer rows.Close()

	for rows.Next() {
		var addr string
		err := rows.Scan(&addr)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs
}

func (d *DefaultDatabase) SaveTokenPrices(tokenPrices types.TokenPrices) {
	for _, tokenPrice := range tokenPrices {
		_, err := d.db.Exec(
			"INSERT INTO token_price (id, public_id, price) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE price = ?",
			tokenPrice.Id,
			tokenPrice.PublicId,
			tokenPrice.Price,
			tokenPrice.Price,
		)
		if err != nil {
			log.Error("Cannot insert into db, token = ", tokenPrice, " err = ", err)
		}
	}
}

func (d *DefaultDatabase) LoadPrices() types.TokenPrices {
	prices := make([]*types.TokenPrice, 0)

	rows, err := d.db.Query("SELECT id, public_id, price FROM token_price")
	if err != nil {
		log.Error("Cannot load prices")
		return prices
	}

	for rows.Next() {
		var NullableId, NullablePublicId sql.NullString
		var price float32

		rows.Scan(&NullableId, &NullablePublicId, &price)

		prices = append(prices, &types.TokenPrice{
			Id:       NullableId.String,
			PublicId: NullablePublicId.String,
			Price:    price,
		})
	}

	defer rows.Close()

	return prices
}

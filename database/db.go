package database

import (
	"database/sql"
	"fmt"
	"math/big"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"

	_ "github.com/mattn/go-sqlite3"
)

// go:generate mockgen -source database/db.go -destination=tests/mock/database/db.go -package=mock
type Database interface {
	Init() error
	SaveTxs(chain string, blockHeight int64, txs *types.Txs)

	// Vault address
	SetVault(chain, address string, token string) error
	GetVaults(chain string) ([]string, error)

	// Token price
	SaveTokenPrices(tokenPrices []*types.TokenPrice)
	LoadPrices() []*types.TokenPrice
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

	// // Connect to the postgres db
	url := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, username, password,
	)
	database, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}

	err = d.createDeyesTables(database, schema)
	if err != nil {
		return err
	}

	d.db = database
	log.Info("Db is connected successfully")
	return nil
}

func (d *DefaultDatabase) createDeyesTables(database *sql.DB, schema string) error {
	rows, err := database.Query("SELECT FROM pg_catalog.pg_database WHERE datname = $1", schema)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		log.Infof("CREATING %s DATABASE", schema)
		// Postgres does not allow using params for table name. Make sure that this schema is safe to
		// pass in.
		_, err := database.Exec("CREATE DATABASE " + schema)
		if err != nil {
			return err
		}
	} else {
		log.Infof("The schema %s already existed", schema)
	}

	return nil
}

func (d *DefaultDatabase) doSqlMigration() error {
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
	if err != nil {
		return err
	}

	// Write the migrations to a temporary directory
	// so they don't need to be managed out of band from the dheart binary.
	migrationDir, err := MigrationsTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for migrations: %w", err)
	}
	defer os.RemoveAll(migrationDir)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationDir,
		"postgres",
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

	err = d.doSqlMigration()
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

		_, err := d.db.Exec("INSERT INTO transactions (chain, tx_hash, block_height, tx_bytes) VALUES (?, ?, ?, ?)",
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

func (d *DefaultDatabase) SetVault(chain, address string, token string) error {
	return d.addWatchAddress(chain, address, fmt.Sprintf("vault__%s", token))
}

func (d *DefaultDatabase) addWatchAddress(chain, address, typ string) error {
	var err error
	if d.cfg.InMemory {
		_, err = d.db.Exec("INSERT INTO watch_address (chain, address, type) VALUES (?, ?, ?)", chain, address, typ)
	} else {
		_, err = d.db.Exec("INSERT INTO watch_address (chain, address, type) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE address=?", chain, address, typ, address)
	}
	if err != nil {
		log.Error(fmt.Sprintf("cannot insert watch address with chain %s and address %s.", chain, address), "Err =", err)
	}

	return err
}

func (d *DefaultDatabase) GetVaults(chain string) ([]string, error) {
	return d.getWatchAddress(chain, "vault__")
}

func (d *DefaultDatabase) getWatchAddress(chain, typ string) ([]string, error) {
	rows, err := d.db.Query("SELECT address FROM watch_address WHERE chain=? and type LIKE ?", chain, typ)
	if err != nil {
		log.Error("Failed to load watch address for chain ", chain, ". Error = ", err)
		return nil, err
	}

	defer rows.Close()
	ret := make([]string, 0)

	if rows.Next() {
		var addr sql.NullString
		err = rows.Scan(&addr)

		if err != nil {
			return nil, err
		}

		ret = append(ret, addr.String)
	}

	return ret, nil
}

func (d *DefaultDatabase) SaveTokenPrices(tokenPrices []*types.TokenPrice) {
	for _, tokenPrice := range tokenPrices {
		_, err := d.db.Exec(
			"INSERT INTO token_price (id, public_id, price) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET price = ?",
			tokenPrice.Id,
			tokenPrice.PublicId,
			tokenPrice.Price.String(),
			tokenPrice.Price.String(),
		)
		if err != nil {
			log.Error("Cannot insert into db, token = ", tokenPrice, " err = ", err)
		}
	}
}

func (d *DefaultDatabase) LoadPrices() []*types.TokenPrice {
	prices := make([]*types.TokenPrice, 0)

	rows, err := d.db.Query("SELECT id, public_id, price FROM token_price")
	if err != nil {
		log.Error("Cannot load prices")
		return prices
	}

	for rows.Next() {
		var nullablePrice, nullableId, nullablePublicId sql.NullString
		rows.Scan(&nullableId, &nullablePublicId, &nullablePrice)

		price, ok := new(big.Int).SetString(nullablePrice.String, 10)
		if !ok {
			return make([]*types.TokenPrice, 0)
		}

		prices = append(prices, &types.TokenPrice{
			Id:       nullableId.String,
			PublicId: nullablePublicId.String,
			Price:    price,
		})
	}

	defer rows.Close()

	return prices
}

// dropSchema drops a table. This should be only be used in unit test to reset the embedded db.
func (d *DefaultDatabase) dropSchema(schema string) error {
	// Postgres does not allow using params for table name. Make sure that this schema is safe to
	// pass in.
	_, err := d.db.Exec("DROP DATABASE " + schema)
	return err
}

package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
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

	// Gateway address
	SetGateway(chain, address string) error
	GetGateway(chain string) (string, error)

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

	var database *sql.DB
	var err error
	if !d.cfg.InMemory {
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
	}

	if d.cfg.InMemory {
		database, err = sql.Open("sqlite3", ":memory:")
		if err != nil {
			return err
		}
	} else {
		database, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, schema))
		if err != nil {
			return err
		}
	}

	d.db = database
	log.Info("Db is connected successfully")
	return nil
}

func (d *DefaultDatabase) doSqlMigration() error {
	driver, err := mysql.WithInstance(d.db, &mysql.Config{})
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

// inMemoryMigration does sql migration for in-memory db. We manually do migration instead of using
// "golang-migrate/migrate" lib because there are some query in "golang-migrate/migrate" not
// supported by sqlite3 in-memory (like SELECT DATABASE() or SELECT GET_LOCK()).
func (d *DefaultDatabase) inMemoryMigration() error {
	log.Verbose("Running in-memory migration...")

	migrationDir, err := MigrationsTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for migrations: %w", err)
	}
	defer os.RemoveAll(migrationDir)

	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return err
	}

	migrationFiles := make([]string, 0)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}

	// Read query from the migration files and execute.
	sort.Strings(migrationFiles)
	for _, f := range migrationFiles {
		dat, err := os.ReadFile(filepath.Join(migrationDir, f))
		if err != nil {
			return err
		}
		query := string(dat)

		_, err = d.db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DefaultDatabase) Init() error {
	err := d.Connect()
	if err != nil {
		log.Error("Failed to connect to DB. Err =", err)
		return err
	}

	if d.cfg.InMemory {
		err = d.inMemoryMigration()
	} else {
		err = d.doSqlMigration()
	}

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

func (d *DefaultDatabase) SetGateway(chain, address string) error {
	_, err := d.db.Exec("INSERT OR REPLACE INTO watch_address (chain, address, type) VALUES (?, ?, ?)", chain, address, "gateway")
	if err != nil {
		log.Error(fmt.Sprintf("cannot insert watch address with chain %s and address %s.", chain, address), "Err =", err)
	}

	return err
}

func (d *DefaultDatabase) GetGateway(chain string) (string, error) {
	rows, err := d.db.Query("SELECT address FROM watch_address WHERE chain=?", chain)
	if err != nil {
		log.Error("Failed to load watch address for chain ", chain, ". Error = ", err)
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		var addr sql.NullString
		err = rows.Scan(&addr)

		if err != nil {
			return "", err
		}

		return addr.String, nil
	}

	return "", nil
}

func (d *DefaultDatabase) SaveTokenPrices(tokenPrices []*types.TokenPrice) {
	for _, tokenPrice := range tokenPrices {
		_, err := d.db.Exec(
			"INSERT INTO token_price (id, public_id, price) VALUES (?, ?, ?) ON CONFLICT(id) DO UPDATE SET price = ?",
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

func (d *DefaultDatabase) LoadPrices() []*types.TokenPrice {
	prices := make([]*types.TokenPrice, 0)

	rows, err := d.db.Query("SELECT id, public_id, price FROM token_price")
	if err != nil {
		log.Error("Cannot load prices")
		return prices
	}

	for rows.Next() {
		var nullableId, nullablePublicId sql.NullString
		var price float64
		rows.Scan(&nullableId, &nullablePublicId, &price)

		prices = append(prices, &types.TokenPrice{
			Id:       nullableId.String,
			PublicId: nullablePublicId.String,
			Price:    price,
		})
	}

	defer rows.Close()

	return prices
}

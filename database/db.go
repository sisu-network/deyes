package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
)

type Database struct {
	db *sql.DB
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
	return &Database{}
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
	fmt.Println("Db is connected successfully")
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

func (d *Database) SaveTx(chain string, hash string, blockHeight int64, bytes []byte) error {
	if len(hash) > 256 {
		hash = hash[:256]
	}

	_, err := d.db.Exec("INSERT INTO transactions (chain, tx_hash, block_height, tx_bytes) VALUES (?, ?, ?, ?)", chain, hash, blockHeight, bytes)
	if err != nil {
		return err
	}

	return nil
}

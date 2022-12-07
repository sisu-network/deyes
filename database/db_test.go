package database

import (
	"fmt"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/jmoiron/sqlx"

	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DbTestSuite struct {
	suite.Suite
}

func connect() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable")
	return db, err
}

func (suite *DbTestSuite) SetupTest() {
	database := embeddedpostgres.NewDatabase()

	if err := database.Start(); err != nil {
		panic(err)
	}

	defer func() {
		if err := database.Stop(); err != nil {
			panic(err)
		}
	}()

	db, err := connect()
	if err != nil {
		panic(err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		panic(err)
	}

	m.Log = &dbLogger{}
	m.Up()
}

func getTestDb(t *testing.T) Database {
	cfg := config.Deyes{
		DbHost:   "127.0.0.1",
		DbSchema: "deyes",
		InMemory: true,
	}
	dbInstance := NewDb(&cfg)
	err := dbInstance.Init()
	require.Nil(t, err)

	return dbInstance
}

func (suite *DbTestSuite) TestDefaultDatabase_SetGateway() {
	fmt.Println("Your First test")
}

func (suite *DbTestSuite) Test_Two() {
	fmt.Println("Your SEcond Test")
}

func (suite *DbTestSuite) Test_Thress() {
	fmt.Println("Your Third Test")
}

func TestDbSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}

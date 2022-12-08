package database

import (
	"fmt"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"

	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	TestSchema = "deyes"
)

// To run a single test in a test suite, specify the -testify.m option.
type DbTestSuite struct {
	suite.Suite
	embeddedDb *embeddedpostgres.EmbeddedPostgres
	sisuDb     Database
}

func (suite *DbTestSuite) SetupTest() {
	host := "localhost"
	port := 5431
	username := "postgres"
	password := "postgres"
	schema := TestSchema

	// Start the embedded server
	dbConfig := embeddedpostgres.DefaultConfig()
	dbConfig = dbConfig.Port(uint32(port)).Username(username).Password(password).Database("postgres")
	suite.embeddedDb = embeddedpostgres.NewDatabase(dbConfig)

	if err := suite.embeddedDb.Start(); err != nil {
		fmt.Println("Err = ", err)
		panic(err)
	}

	deyesCfg := &config.Deyes{
		DbHost:     host,
		DbPort:     port,
		DbUsername: username,
		DbPassword: password,
		DbSchema:   schema,
	}

	db := NewDb(deyesCfg)
	err := db.Init()
	if err != nil {
		panic(err)
	}

	suite.sisuDb = db
}

func (suite *DbTestSuite) TearDownTest() {
	fmt.Println("Tearing down test....")
	if suite.sisuDb != nil {
		err := suite.sisuDb.(*DefaultDatabase).dropSchema(TestSchema)
		if err != nil {
			panic(err)
		}
	}

	if err := suite.embeddedDb.Stop(); err != nil {
		panic(err)
	}
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

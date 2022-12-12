package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationDbSuite struct {
	suite.Suite
}

func resetDb() {
	cfg := getTestDbConfig()
	db := NewDb(&cfg).(*DefaultDatabase)
	db.Init()
	db.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", DbSchema))
	db.Close()
}

func (suite *IntegrationDbSuite) TestSetVaults() {
	resetDb()
	testSetVaults(suite.T(), false)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationDbSuite))
}

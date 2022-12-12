package database

import (
	"testing"

	"github.com/sisu-network/deyes/config"
)

func getTestDbConfig() config.Deyes {
	cfg := config.Deyes{
		DbHost:     "localhost",
		DbPort:     3306,
		DbUsername: "root",
		DbPassword: "password",
		DbSchema:   "TestDb",
	}

	return cfg
}

func TestInMemory_SetVaults(t *testing.T) {
	testSetVaults(t, true)
}

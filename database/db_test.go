package database

import (
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"
)

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

func TestDefaultDatabase_SetGateway(t *testing.T) {
	db := getTestDb(t)

	err := db.SetGateway("eth", "addr0")
	require.Nil(t, err)

	addr, err := db.GetGateway("eth")
	require.Nil(t, err)
	require.Equal(t, "addr0", addr)
}

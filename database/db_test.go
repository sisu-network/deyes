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

func TestDefaultDatabase_SaveWatchAddress(t *testing.T) {
	db := getTestDb(t)

	err := db.SaveWatchAddress("eth", "addr0")
	require.Nil(t, err)

	addrs := db.LoadWatchAddresses("eth")
	require.Equal(t, 1, len(addrs))

	err = db.SaveWatchAddress("eth", "addr1")
	require.Nil(t, err)

	addrs = db.LoadWatchAddresses("eth")
	require.Equal(t, 2, len(addrs))
	require.Equal(t, "addr0", addrs[0].Address)
	require.Equal(t, "addr1", addrs[1].Address)
}

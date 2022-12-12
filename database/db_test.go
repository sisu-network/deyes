package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testSetVaults(t *testing.T, inMemory bool) {
	cfg := getTestDbConfig()
	cfg.InMemory = inMemory
	db := NewDb(&cfg)
	err := db.Init()
	require.Nil(t, err)

	chain1 := "ganache1"
	db.SetVault(chain1, "addr1", "")
	chain2 := "ganache2"
	db.SetVault(chain2, "addr2", "")

	chain3 := "ganache3"
	db.SetVault(chain3, "addr3_1", "token1")
	db.SetVault(chain3, "addr3_2", "token2")

	vault1, err := db.GetVaults(chain1)
	require.Nil(t, err)
	require.Equal(t, []string{"addr1"}, vault1)

	vault2, err := db.GetVaults(chain2)
	require.Nil(t, err)
	require.Equal(t, []string{"addr2"}, vault2)

	vault3, err := db.GetVaults(chain3)
	require.Nil(t, err)
	require.Equal(t, []string{"addr3_1", "addr3_2"}, vault3)

	// Update the token1 address with new address
	db.SetVault(chain3, "addr3_3", "token1")
	vault3, err = db.GetVaults(chain3)
	require.Nil(t, err)
	require.Equal(t, []string{"addr3_3", "addr3_2"}, vault3)
}

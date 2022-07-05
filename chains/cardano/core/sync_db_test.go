package core

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSyncDB(t *testing.T) {
	t.Skip()
	t.Parallel()

	cfg := PostgresConfig{
		Host:     "hide",
		Port:     3000,
		User:     "hide",
		Password: "hide",
		DbName:   "cexplorer",
	}

	db, err := connectDB(cfg)
	require.NoError(t, err)

	syncDB := NewSyncDBConnector(db)
	syncDB.TransactionUTXOs(context.Background(), "\\xd02ef28bf0e05876d6d80b16993fdc9805402247691e24d0caed7c4c32bbe9f4")
}

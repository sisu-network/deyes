package core

import (
	"context"
	"fmt"
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
	meta, err := syncDB.TransactionMetadata(context.Background(), `\x1a2c7de4efa266d52dae95454c5671a47f45742b0fceb012a791755c3a75c2fc`)
	require.NoError(t, err)
	fmt.Println(meta)
}

package core

import (
	"context"
	"fmt"
	"github.com/sisu-network/deyes/config"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSyncDB(t *testing.T) {
	t.Skip()

	t.Parallel()

	cfg := config.SyncDbConfig{
		Host:     "hide",
		Port:     5432,
		User:     "hide",
		Password: "hide",
		DbName:   "hide",
	}

	db, err := ConnectDB(cfg)
	require.NoError(t, err)

	syncDB := NewSyncDBConnector(db)
	utxos, err := syncDB.TransactionUTXOs(context.Background(), "e61fbf283d12890ff50e9e466175573209773b4abe58b7c33e1449ba08c02a74")
	require.NoError(t, err)
	for _, o := range utxos.Outputs {
		fmt.Println(o.Address)
	}

	metadata, err := syncDB.TransactionMetadata(context.Background(), "ded242b05f506e870681a780562fa39fdab4bfd0a24fb275a3e4b4ef0a4d15a1")
	for _, m := range metadata {
		fmt.Println(m.JsonMetadata)
	}
}

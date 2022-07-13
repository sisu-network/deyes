package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/blockfrost/blockfrost-go"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSyncDB(t *testing.T) {
	//t.Skip()

	t.Parallel()

	cfg := PostgresConfig{
		Host:     "143.198.98.1",
		Port:     5432,
		User:     "sisu",
		Password: "sisu",
		DbName:   "cexplorer",
	}

	db, err := ConnectDB(cfg)
	require.NoError(t, err)

	syncDB := NewSyncDBConnector(db)
	//meta, err := syncDB.TransactionMetadata(context.Background(), `\x1a2c7de4efa266d52dae95454c5671a47f45742b0fceb012a791755c3a75c2fc`)
	//require.NoError(t, err)
	//fmt.Println(meta)

	txs, err := syncDB.AddressTransactions(context.Background(), "addr_test1qrj8mcevhx4s7q7uxe9yx6fsl3e5vxcshl9j5kvj33n2gmrdh0u8y9amknkkkuqd8rxf9yanp7dexxw0w52c9rqmlz7swaf0ur", blockfrost.APIQueryParams{From: "3702442"})
	require.NoError(t, err)
	fmt.Println(txs)
}

package core

import (
	"context"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSyncDB(t *testing.T) {
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
	utxos, err := syncDB.TransactionUTXOs(context.Background(), "\\x97228723b810ec31064a3b7bcd83138301295bacaa62385cd930774e8901fc68")
	require.NoError(t, err)
	for _, output := range utxos.Outputs {
		fmt.Println("address = ", output.Address)
		for _, amt := range output.Amount {
			fmt.Println("unit = ", amt.Unit, " quantity = ", amt.Quantity)
		}
	}
}

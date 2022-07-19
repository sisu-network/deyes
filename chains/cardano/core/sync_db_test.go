package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/blockfrost/blockfrost-go"
	"github.com/sisu-network/deyes/config"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func TestIntegrationSyncDB(t *testing.T) {
	t.Skip()

	t.Parallel()

	cfg := config.SyncDbConfig{}

	db, err := ConnectDB(cfg)
	require.NoError(t, err)

	syncDB := NewSyncDBConnector(db)
	utxos, err := syncDB.AddressUTXOs(context.Background(), "addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy", blockfrost.APIQueryParams{To: "3719135"})
	require.NoError(t, err)
	for _, u := range utxos {
		fmt.Printf("%+v\n", u)
	}
}

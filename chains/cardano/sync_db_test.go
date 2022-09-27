package cardano

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

func TestIntegrationSyncDB(t *testing.T) {
	t.Skip()

	t.Parallel()

	cfg := config.SyncDbConfig{}

	str, err := json.Marshal(&cfg)
	require.NoError(t, err)
	fmt.Println(string(str))

	db, err := ConnectDB(cfg)
	require.NoError(t, err)

	syncDB := NewSyncDBConnector(db)
	utxos, err := syncDB.AddressUTXOs(context.Background(), "addr_test1vrfdtdcy8tu8000jprfclp8dz9d6pgl2984fvtzhnqafx7qmmg0l4",
		types.APIQueryParams{To: "18446744073709551615"})
	require.NoError(t, err)
	require.NotEmpty(t, utxos)
}

func TestBuildQueryFromString(t *testing.T) {
	t.Parallel()

	arr := []int64{1, 2, 3, 4}
	str := buildQueryFromIntArray(arr)
	require.Equal(t, "(1,2,3,4)", str)
}

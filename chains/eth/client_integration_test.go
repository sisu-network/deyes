package eth

import (
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestIntegration_GetExtraRpcs(t *testing.T) {
	t.Skip()

	c := NewEthClients([]string{}, config.Chain{Chain: "goerli-testnet"}, true).(*defaultEthClient)
	rpcs, err := c.GetExtraRpcs()
	require.Nil(t, err)

	filterRpcs, _, _ := c.getRpcsHealthiness(rpcs)
	log.Verbose("filterRpcs = ", filterRpcs)
}

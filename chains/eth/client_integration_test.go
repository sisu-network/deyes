package eth

import (
	"testing"

	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/lib/log"
	"github.com/stretchr/testify/require"
)

func TestIntegration_GetExtraRpcs(t *testing.T) {
	t.Skip()

	c := NewEthClients(
		config.Chain{
			Chain:         "goerli-testnet",
			UseGasEip1559: true,
			Rpcs:          make([]string, 0),
		},
		true,
	).(*defaultEthClient)
	rpcs, err := c.GetExtraRpcs()
	require.Nil(t, err)

	fileterRpcs, _, _ := c.getRpcsHealthiness(rpcs)
	log.Verbose("fileterRpcs = ", fileterRpcs)
}

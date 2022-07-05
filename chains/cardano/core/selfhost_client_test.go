package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegrationSelfHostClient(t *testing.T) {
	t.Skip()

	t.Parallel()

	cfg := PostgresConfig{
		Host:     "hide",
		Port:     3000,
		User:     "hide",
		Password: "hide",
		DbName:   "cexplorer",
	}

	client := NewSelfHostClient(cfg, "")
	h, err := client.BlockHeight()
	require.NoError(t, err)

	fmt.Println("height = ", h)
}

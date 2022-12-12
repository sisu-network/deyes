package database

import (
	"math/big"
	"testing"

	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

// This file contains integration test for sanity checking
func TestTokenPrice(t *testing.T) {
	t.Skip()

	cfg := getTestDbConfig()

	db := NewDb(&cfg)
	err := db.Init()
	require.Nil(t, err)

	tokenPrices := []*types.TokenPrice{
		{
			Id:       "ETH",
			PublicId: "ETH",
			Price:    big.NewInt(10000),
		},
	}

	db.SaveTokenPrices(tokenPrices)
}

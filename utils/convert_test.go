package utils

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloatToWei(t *testing.T) {
	value := float64(1.234)
	conv := FloatToWei(value)

	expected, _ := new(big.Int).SetString("1234000000000000000", 10)
	require.Equal(t, expected, conv)
}

func TestToSisuPrice(t *testing.T) {
	price := big.NewInt(6 * 100_000_000)
	converted := ToSisuPrice(price, 8)
	require.Equal(t, big.NewInt(SisuUnit*6), converted)
}

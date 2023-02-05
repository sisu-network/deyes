package oracle

import (
	"math/big"

	"github.com/sisu-network/deyes/config"
)

type Provider interface {
	GetPrice(token config.Token) (*big.Int, error)
}

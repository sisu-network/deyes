package types

import "math/big"

type TokenPrice struct {
	Id       string
	PublicId string
	Price    *big.Int
}

package types

import (
	"math/big"
)

type TransferOutData struct {
	Amount       big.Int
	TokenAddress string
	ChainId      uint64
	Recipient    string
}

func NewTransferOutData(amount *big.Int) *TransferOutData {
	return &TransferOutData{}
}

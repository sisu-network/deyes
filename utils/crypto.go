package utils

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
)

func PublicKeyBytesToAddress(publicKey []byte) common.Address {
	var buf []byte

	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKey[1:]) // remove EC prefix 04
	buf = hash.Sum(nil)
	address := buf[12:]

	return common.HexToAddress(hex.EncodeToString(address))
}

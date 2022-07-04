package utils

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// Hash a string and return the first 32 bytes of the hash.
func KeccakHash32(s string) string {
	return KeccakHash32Bytes([]byte(s))
}

func KeccakHash32Bytes(bz []byte) string {
	hash := sha3.NewLegacyKeccak256()

	var buf []byte
	hash.Write(bz)
	buf = hash.Sum(nil)

	encoded := hex.EncodeToString(buf)
	if len(encoded) > 32 {
		encoded = encoded[:32]
	}

	return encoded
}

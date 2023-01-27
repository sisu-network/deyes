package types

import (
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/protobuf/proto"
)

func (x *TransactionMessage) GetHash() (string, error) {
	bz, err := proto.Marshal(x)
	if err != nil {
		return "", err
	}

	sha := sha256.Sum256(bz)
	return hex.EncodeToString(sha[:]), nil
}

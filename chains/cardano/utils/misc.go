package utils

import "github.com/echovl/cardano-go"

func GetAddressFromCardanoPubkey(pubkey []byte) cardano.Address {
	keyHash, err := cardano.Blake224Hash(pubkey)
	if err != nil {
		panic(err)
	}

	payment := cardano.StakeCredential{Type: cardano.KeyCredential, KeyHash: keyHash}
	enterpriseAddr, err := cardano.NewEnterpriseAddress(0, payment)
	if err != nil {
		panic(err)
	}

	return enterpriseAddr
}

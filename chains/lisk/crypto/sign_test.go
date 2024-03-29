package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	defaultMessage      = "Some default text."
	signPublicKey, _    = hex.DecodeString("7ef45cd525e95b7a86244bbd4eb4550914ad06301013958f4dd64d32ef7bc588")
	signPrivateKey, _   = hex.DecodeString("314852d7afb0d4c283692fef8a2cb40e30c7a5df2ed79994178c10ac168d6d977ef45cd525e95b7a86244bbd4eb4550914ad06301013958f4dd64d32ef7bc588")
	defaultSignature, _ = hex.DecodeString("974eeac2c7e7d9da42aa273c8caae8e6eb766fa29a31b37732f22e6d2e61e8402106849e61e3551ff70d7d359170a6198669e1061b6b4aa61997e26b87e3a704")
	wrongSignature, _   = hex.DecodeString("974f2ac2c7e7d9da42aa273c8caae8e6eb766fa29a31b37732f22e6d2e61e8402106849e61e3551ff70d7d359170a6198669e1061b6b4aa61997e26b87e3a704")
)

func TestSign_MessageWithPrivateKey(t *testing.T) {
	val := SignMessage([]byte(defaultMessage), signPrivateKey)
	require.Equal(t, val, defaultSignature)
}

func TestSign_DataWithPrivateKey(t *testing.T) {
	val := SignMessage([]byte(defaultMessage), signPrivateKey)
	require.Equal(t, val, defaultSignature)
}

func TestVerify_DataWithPublicKey(t *testing.T) {
	//  verify Message With valid publicKey
	isVerifiedMessage := VerifyMessage([]byte(defaultMessage), defaultSignature, signPublicKey)
	require.Equal(t, isVerifiedMessage, true)

	//  verify Message With invalid publicKey
	messageVal := VerifyMessage([]byte(defaultMessage), wrongSignature, signPublicKey)
	require.Equal(t, messageVal, false)

	//  verify Data With valid publicKey
	isVerifiedData := VerifyMessage([]byte(defaultMessage), defaultSignature, signPublicKey)
	require.Equal(t, isVerifiedData, true)

	//  verify Data With invalid publicKey
	dataVal := VerifyMessage([]byte(defaultMessage), wrongSignature, signPublicKey)
	require.Equal(t, dataVal, false)
}

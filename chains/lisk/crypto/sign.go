package crypto

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/ed25519"
)

// SignMessageWithPrivateKey takes a message and a privateKey and returns a signature as hex string
func SignMessageWithPrivateKey(rawMessage []byte, privateKey []byte) []byte {
	signedMessage := ed25519.Sign(privateKey, rawMessage)

	return signedMessage
}

// SignDataWithPrivateKey takes data and a privateKey and returns a signature
func SignDataWithPrivateKey(data []byte, privateKey []byte) []byte {
	signedMessage := ed25519.Sign(privateKey, data)

	return signedMessage
}

// VerifyMessageWithPublicKey takes a message, signature and publicKey and verifies it
func VerifyMessageWithPublicKey(message string, signature []byte, publicKey []byte) bool {
	isValid := ed25519.Verify(publicKey, []byte(message), signature)

	return isValid
}

// VerifyDataWithPublicKey takes data, a signature and a publicKey and verifies it
func VerifyDataWithPublicKey(data []byte, signature []byte, publicKey []byte) bool {
	isValid := ed25519.Verify(publicKey, data, signature)

	return isValid
}

func SignWithNetwork(network string, txBytes []byte, privateKey []byte) ([]byte, error) {
	bz, err := GetSigningBytes(network, txBytes, privateKey)
	if err != nil {
		return nil, err
	}

	return SignMessageWithPrivateKey(bz, privateKey), nil
}

func GetSigningBytes(network string, txBytes []byte, privateKey []byte) ([]byte, error) {
	dst := new(bytes.Buffer)
	//First byte is the network info
	networkBytes, err := hex.DecodeString(network)
	if err != nil {
		return nil, err
	}

	binary.Write(dst, binary.LittleEndian, networkBytes)

	// Append the transaction ModuleID
	binary.Write(dst, binary.LittleEndian, txBytes)

	return dst.Bytes(), nil
}

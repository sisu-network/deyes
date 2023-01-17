package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/ed25519"
)

var (
	GENERATOR = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	CHARSET   = "zxvcpmbn3465o978uyrtkqew2adsjhfg"
)

// GetPrivateKeyFromSecret takes a Lisk secret and returns the associated private key
func GetPrivateKeyFromSecret(secret string) []byte {
	secretHash := GetSHA256Hash(secret)
	_, prKey, _ := ed25519.GenerateKey(bytes.NewReader(secretHash[:sha256.Size]))

	return prKey
}

// GetPublicKeyFromSecret takes a Lisk secret and returns the associated public key
func GetPublicKeyFromSecret(secret string) []byte {
	secretHash := GetSHA256Hash(secret)
	pKey, _, _ := ed25519.GenerateKey(bytes.NewReader(secretHash[:sha256.Size]))

	return pKey
}

// GetAddressFromPublicKey takes a Lisk public key and returns the associated address
func GetAddressFromPublicKey(publicKey []byte) string {
	publicKeyHash := sha256.Sum256(publicKey)
	return hex.EncodeToString(publicKeyHash[:20])
}

// GetLisk32AddressFromPublickey returns a Lisk 32 bytes format from public key.
func GetLisk32AddressFromPublickey(publicKey []byte) string {
	publicKeyHash := sha256.Sum256(publicKey)
	addrBytes := publicKeyHash[:20]
	return AddressToLisk32(addrBytes)
}

func AddressToLisk32(address []byte) string {
	var byteSequence []byte
	for _, b := range address {
		byteSequence = append(byteSequence, b)
	}
	uint5Address := ConvertUIntArray(byteSequence, 8, 5)
	uint5Checksum := CreateChecksum(uint5Address)

	return "lsk" + ConvertUInt5ToBase32(append(uint5Address, uint5Checksum...))
}

func ConvertUIntArray(uintArray []byte, fromBits int, toBits int) []byte {
	maxValue := (1 << toBits) - 1
	accumulator := 0
	bits := 0

	var result []byte
	for _, p := range uintArray {
		byteValue := p
		if byteValue < 0 || byteValue>>fromBits != 0 {
			return make([]byte, 0)
		}
		accumulator = (accumulator << fromBits) | int(byteValue)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((accumulator>>bits)&maxValue))
		}
	}
	return result
}

func CreateChecksum(uint5Array []byte) []byte {
	values := append(uint5Array, []byte{0, 0, 0, 0, 0, 0}...)
	mod := Polymod(values) ^ 1
	var result []byte
	for p := 0; p < 6; p += 1 {
		result = append(result, byte((mod>>(5*(5-p)))&31))
	}
	return result
}

func Polymod(uint5Array []byte) int {
	chk := 1
	for _, value := range uint5Array {

		top := chk >> 25
		chk = ((chk & 0x1ffffff) << 5) ^ int(value)
		for i := 0; i < 5; i += 1 {
			if ((top >> i) & 1) > 0 {
				chk ^= GENERATOR[i]
			}
		}
	}
	return chk
}

func ConvertUInt5ToBase32(uint5Array []byte) string {
	result := ""
	charsets := strings.Split(CHARSET, "")
	for _, value := range uint5Array {
		result += charsets[value]
	}
	return result
}

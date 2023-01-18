package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/ed25519"
)

var (
	GENERATOR             = []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	CHARSET               = "zxvcpmbn3465o978uyrtkqew2adsjhfg"
	LISK32_ADDRESS_LENGTH = 41
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

// Lisk32AddressToPublicAddress converts lisk32 address to a public address (byte array) used in a
// transaction. For example:
// lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pvf -> dc f5 7d 8b f3 3b b4 6b 51 f8 ec b9 1b 78 ea 45 3d 95 31 4d
func Lisk32AddressToPublicAddress(lisk32 string) ([]byte, error) {
	err := ValidateLisk32(lisk32)
	if err != nil {
		return nil, err
	}

	// Base32 with no prefix and checksum
	base32 := lisk32[3 : len(lisk32)-6]
	intSeq := make([]byte, 0)
	for _, r := range base32 {
		index := strings.IndexRune(CHARSET, r)
		intSeq = append(intSeq, byte(index))
	}

	intSeq8 := ConvertUIntArray(intSeq, 5, 8)

	return intSeq8, nil
}

func ValidateLisk32(lisk32 string) error {
	if len(lisk32) != LISK32_ADDRESS_LENGTH {
		return fmt.Errorf("invalid lisk address length, lisk32 = %s", lisk32)
	}

	if lisk32[:3] != "lsk" {
		return fmt.Errorf("invalid lisk prefix, lisk32 = %s", lisk32)
	}

	subAddr := lisk32[3:]

	integerSeq := make([]byte, 0)
	for _, r := range subAddr {
		index := strings.IndexRune(CHARSET, r)
		if index < 0 {
			return fmt.Errorf("invalid lisk character %v", r)
		}

		integerSeq = append(integerSeq, byte(index))
	}

	// Verify checksum
	if Polymod(integerSeq) != 1 {
		return fmt.Errorf("check sum failed")
	}

	return nil
}

package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	passphrase = "camera accident escape cricket frog pony record occur broken inhale waste swing"
)

func TestGet_PublicKeyFromSecret(t *testing.T) {
	publicKey, _ := hex.DecodeString("f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	val := GetPublicKeyFromSecret(passphrase)
	require.Equal(t, val, publicKey)
}

func TestGet_PrivateKeyFromSecret(t *testing.T) {
	privateKey, _ := hex.DecodeString("ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")

	val := GetPrivateKeyFromSecret(passphrase)
	require.Equal(t, val, privateKey)
}

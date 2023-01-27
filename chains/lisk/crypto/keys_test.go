package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	passphrase = "camera accident escape cricket frog pony record occur broken inhale waste swing"
)

func TestKey_PublicKeyFromSecret(t *testing.T) {
	publicKey, _ := hex.DecodeString("f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	val := GetPublicKeyFromSecret(passphrase)
	require.Equal(t, val, publicKey)
}

func TestKey_PrivateKeyFromSecret(t *testing.T) {
	privateKey, _ := hex.DecodeString("ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")

	val := GetPrivateKeyFromSecret(passphrase)
	require.Equal(t, val, privateKey)
}

func TestKey_ValidateLisk32(t *testing.T) {
	err := ValidateLisk32("lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pvf")
	require.Nil(t, err)

	err = ValidateLisk32("lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pv")
	require.NotNil(t, err)
	require.Equal(t,
		"invalid lisk address length, lisk32 = lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pv",
		err.Error(),
	)

	err = ValidateLisk32("lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pvu")
	require.NotNil(t, err)
	require.Equal(t, "check sum failed", err.Error())
}

func TestKey_Lisk32AddressToPublicAddress(t *testing.T) {
	pubAddr, err := Lisk32AddressToPublicAddress("lskstdwsvgtn7dbek82hrjyefn63kfa6o69uz9pvf")
	require.Nil(t, err)

	require.Equal(t, "dcf57d8bf33bb46b51f8ecb91b78ea453d95314d", hex.EncodeToString(pubAddr))
}

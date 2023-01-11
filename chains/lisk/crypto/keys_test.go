package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var (
	passphraseHash = "ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76"
	passphrase     = "camera accident escape cricket frog pony record occur broken inhale waste swing"
	privateKey, _  = hex.DecodeString("ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	publicKey, _   = hex.DecodeString("f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	address        = "lsk9hxtj8busjfugaxcg9zfuzdty7zyagcrsxvohk"
)

func TestGetPublicKeyFromSecret(t *testing.T) {
	if val := GetPublicKeyFromSecret(passphrase); !bytes.Equal(val, publicKey) {
		t.Errorf("GetPublicKeyFromSecret(%v)=%v; want %v", passphrase, val, publicKey)
	}
}

func TestGetPrivateKeyFromSecret(t *testing.T) {
	if val := GetPrivateKeyFromSecret(passphrase); !bytes.Equal(val, privateKey) {
		t.Errorf("GetPrivateKeyFromSecret(%v)=%v; want %v", passphrase, val, privateKey)
	}
}

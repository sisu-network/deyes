package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var (
	defaultPassphraseHash = "6f433e3cd0df3d0fe423fb41e803538b811cfc72"
	defaultPassphrase     = "camera accident escape cricket frog pony record occur broken inhale waste swing"
	defaultPrivateKey, _  = hex.DecodeString("ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	defaultPublicKey, _   = hex.DecodeString("f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	defaultAddress        = "lsk9hxtj8busjfugaxcg9zfuzdty7zyagcrsxvohk"
)

func TestGetPublicKeyFromSecret(t *testing.T) {
	if val := GetPublicKeyFromSecret(defaultPassphrase); !bytes.Equal(val, defaultPublicKey) {
		t.Errorf("GetPublicKeyFromSecret(%v)=%v; want %v", defaultPassphrase, val, defaultPublicKey)
	}
}

func TestGetPrivateKeyFromSecret(t *testing.T) {
	if val := GetPrivateKeyFromSecret(defaultPassphrase); !bytes.Equal(val, defaultPrivateKey) {
		t.Errorf("GetPrivateKeyFromSecret(%v)=%v; want %v", defaultPassphrase, val, defaultPrivateKey)
	}
}

package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	defaultBytes                = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	default8BytesReversed       = []byte{7, 6, 5, 4, 3, 2, 1, 0}
	defaultBytesBigNumberString = "18591708106338011145"

	unpaddedData       = []byte{0, 1, 2, 3, 4}
	paddedData         = []byte{0, 1, 2, 3, 4, 2, 2}
	invalidPaddedData  = []byte{0, 1, 2, 3, 4, 2, 20}
	invalidPaddedData2 = []byte{0, 1, 2, 3, 4, 2, 3}
)

func TestGetSHA256Hash(t *testing.T) {
	val := GetSHA256Hash(passphrase)
	require.Equal(t, hex.EncodeToString(val[:]), passphraseHash)

}

func TestGetFirstEightBytesReversed(t *testing.T) {
	successVal := GetFirstEightBytesReversed(defaultBytes)
	require.Equal(t, successVal, default8BytesReversed)

	failedVal := GetFirstEightBytesReversed(nil)
	require.Equal(t, failedVal, []uint8([]byte(nil)))
}

func TestGetBigNumberStringFromBytes(t *testing.T) {
	val := GetBigNumberStringFromBytes(defaultBytes)
	require.Equal(t, val, defaultBytesBigNumberString)
}

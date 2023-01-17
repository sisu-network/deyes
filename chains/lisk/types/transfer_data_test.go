package types

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestTransferData_Serialize(t *testing.T) {
	chainId := int64(1234234233)
	recipient, err := hex.DecodeString("bac265B9e5758F325703bcc6C43F98C84e2F5aD9")
	require.Nil(t, err)

	amount, err := strconv.ParseUint("123124238962348765", 10, 64)
	require.Nil(t, err)

	data := &TransferData{
		ChainId:   chainId,
		Recipient: recipient,
		Amount:    amount,
	}

	bz, err := proto.Marshal(data)
	require.Nil(t, err)

	// Encode
	encoded := base64.StdEncoding.EncodeToString(bz)
	require.True(t, len(encoded) <= 64)

	// Decode
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.Nil(t, err)
	require.Equal(t, bz, decoded)

	tx := &TransferData{}
	err = proto.Unmarshal(decoded, tx)
	require.Nil(t, err)
	require.Equal(t, data, tx)
}

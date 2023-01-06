package lisk_test

import (
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	ltypes "github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestLiskDispatcher_DeserializeTx(t *testing.T) {
	//client := &lisk.MockLiskClient{
	//	GetAccountFunc: func(address string) (*ltypes.Account, error) {
	//		return &ltypes.Account{Summary: &ltypes.AccountSummary{Address: ""}, Sequence: &ltypes.AccountSequence{Nonce: "15"}}, nil
	//	},
	//}
	//acc, _ := client.GetAccount("lsk7kb3tbtrq5t4vbuzbavdahgj7yojmn4qbof63q")

	//moduleId := uint32(2)
	//assetId := uint32(0)
	//nonce, err := strconv.ParseUint(acc.Sequence.Nonce, 10, 32)
	txPb := &ltypes.TransactionMessage{
		//ModuleID:        &moduleId,
		//AssetID:         &assetId,
		//Nonce:           &nonce,
		//Fee:             &nonce,
		//SenderPublicKey: []byte("8f057d088a585d938c20d63e430a068d4cea384e588aa0b758c68fca21644dbc"),
		//Asset:           []byte("f214d75bbc4b2ea89e433f3a45af803725416ec3"),
		//Signatures:      [][]byte{[]byte("204514eb1152355799ece36d17037e5feb4871472c60763bdafe67eb6a38bec632a8e2e62f84a32cf764342a4708a65fbad194e37feec03940f0ff84d3df2a05")},
	}
	bz, err := hex.DecodeString("08021000184c2080c2d72f2a20f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289321e0880c8afa0251214445f5ae1342837a1231f9d36d34a79145c1cd0141a00")
	require.Nil(t, err)

	err = proto.Unmarshal(bz, txPb)
	require.Nil(t, err)
	log.Println(txPb)
	//log.Info(hex.EncodeToString(tx), err)
	// Test serializing & deserializing tx.

}

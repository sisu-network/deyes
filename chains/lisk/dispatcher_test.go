package lisk_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"github.com/sisu-network/deyes/chains/lisk/crypto"
	ltypes "github.com/sisu-network/deyes/chains/lisk/types"
	"golang.org/x/crypto/ed25519"
	"strconv"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/sisu-network/deyes/chains/lisk"
	"github.com/sisu-network/deyes/types"
	"github.com/stretchr/testify/require"
)

var (
	transactionResult    = "08021000180f200f322866323134643735626263346232656138396534333366336134356166383033373235343136656333"
	defaultPassphrase    = "camera accident escape cricket frog pony record occur broken inhale waste swing"
	defaultPrivateKey, _ = hex.DecodeString("ba20a2df297ff5db79764c7b4e778e00eaa81b665b551447fae4fdcd64c81b76f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	defaultPublicKey, _  = hex.DecodeString("f0321539a45078365c1a65944d010876c0efe45c0446101dacced7a2f29aa289")
	defaultAddress       = "lsk9hxtj8busjfugaxcg9zfuzdty7zyagcrsxvohk"
	moduleId             = uint32(2)
	assetId              = uint32(0)
	amount               = uint64(1000000)
	data                 = "test message"
	fee                  = uint64(300000)
	recipientAddress     = "83f03b28f0497eac8aaf6f11dc98e7f733ceb92c"
)

func TestLiskDispatcher_DeserializeTx(t *testing.T) {
	client := &lisk.MockLiskClient{
		GetAccountFunc: func(address string) (*ltypes.Account, error) {
			return &ltypes.Account{Summary: &ltypes.AccountSummary{Address: ""}, Sequence: &ltypes.AccountSequence{Nonce: "15"}}, nil
		},
		CreateTransactionFunc: func(txHash string) (string, error) {
			return transactionResult, nil
		},
	}
	acc, _ := client.GetAccount(defaultAddress)
	nonce, err := strconv.ParseUint(acc.Sequence.Nonce, 10, 32)
	require.Nil(t, err)

	assetPb := &ltypes.AssetMessage{
		Amount:           &amount,
		RecipientAddress: []byte(recipientAddress),
		Data:             &data,
	}
	asset, err := proto.Marshal(assetPb)
	pubKey := crypto.GetPublicKeyFromSecret(defaultPassphrase)
	privateKey := crypto.GetPrivateKeyFromSecret(defaultPassphrase)

	txPb := &ltypes.TransactionMessage{
		ModuleID:        &moduleId,
		AssetID:         &assetId,
		Nonce:           &nonce,
		Fee:             &fee,
		Asset:           asset,
		SenderPublicKey: pubKey,
	}
	signatures, err := sign(txPb, privateKey)
	txPb.Signatures = [][]byte{signatures}
	require.Nil(t, err)
	txHash, err := proto.Marshal(txPb)
	tx := types.DispatchedTxRequest{Chain: "lisk-testnet", TxHash: hex.EncodeToString(txHash)}

	dispatcher := lisk.NewDispatcher("lisk-testnet", client)
	dpResult := dispatcher.Dispatch(&tx)
	require.Equal(t, dpResult.Success, true)

	//test decode transaction with protobuf
	//pb := ltypes.TransactionMessage{}
	//proto.Unmarshal(txHash, &pb)
	//log.Println(pb.GetAsset())
}

func sign(tx *ltypes.TransactionMessage, privateKey []byte) ([]byte, error) {
	dst := new(bytes.Buffer)

	// First byte is the transaction ModuleID
	binary.Write(dst, binary.LittleEndian, tx.ModuleID)

	// Append the AssetID
	binary.Write(dst, binary.LittleEndian, tx.AssetID)

	// Append the Nonce
	binary.Write(dst, binary.LittleEndian, tx.Nonce)

	// Append the Fee
	binary.Write(dst, binary.LittleEndian, tx.Fee)

	// Append the sender's public key
	transactionSenderPubKey := make([]byte, ed25519.PublicKeySize)
	copy(transactionSenderPubKey, tx.SenderPublicKey)
	binary.Write(dst, binary.LittleEndian, transactionSenderPubKey)
	// Append asset data if given
	if tx.Asset != nil {
		binary.Write(dst, binary.LittleEndian, tx.Asset)
	}

	// Append signatures (both optional)
	if len(tx.Signatures) > 0 {
		binary.Write(dst, binary.LittleEndian, tx.Signatures)
	}
	hash := sha256.Sum256(dst.Bytes())
	return crypto.SignMessageWithPrivateKey(string(hash[:]), privateKey), nil
}

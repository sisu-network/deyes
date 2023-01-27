package lisk_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/sisu-network/deyes/chains/lisk/crypto"
	ltypes "github.com/sisu-network/deyes/chains/lisk/types"
	"github.com/sisu-network/deyes/types"

	"github.com/sisu-network/deyes/chains/lisk"
	"github.com/stretchr/testify/require"
)

var (
	transactionResult    = "08021000180f200f322866323134643735626263346232656138396534333366336134356166383033373235343136656333"
	defaultPassphrase    = "limit sort chase funny stumble stove diary regret demand march swim cycle"
	defaultPrivateKey, _ = hex.DecodeString("fb3c93fbc5b28ef0879416d2928cefa8fc456a31f23302776b68f4481a5a1db837c8db9d10cd037f1fcbec7a9c091a86415e71975891b69f23bd60dcba233aa9")
	defaultPublicKey, _  = hex.DecodeString("37c8db9d10cd037f1fcbec7a9c091a86415e71975891b69f23bd60dcba233aa9")
	defaultAddress       = "lskgug634kfkpaatuushjpyvd8cucgfe8afsbsm8u"
	moduleId             = uint32(2)
	assetId              = uint32(0)
	amount               = uint64(1000000000)
	data                 = ""
	fee                  = uint64(100000000)
	recipientAddress, _  = hex.DecodeString("445f5ae1342837a1231f9d36d34a79145c1cd014")
	networks             = map[string]string{"mainnet": "", "testnet": "e8832331820e5ba835012106a7c807b46c8b9c8672b6217b01373773fe87daf8"}
)

func TestLiskDispatcher_DeserializeTx(t *testing.T) {
	// create lisk client
	client := &lisk.MockLiskClient{
		GetAccountFunc: func(address string) (*ltypes.Account, error) {
			return &ltypes.Account{Summary: &ltypes.AccountSummary{Address: ""}, Sequence: &ltypes.AccountSequence{Nonce: "44"}}, nil
		},
		CreateTransactionFunc: func(txHash string) (string, error) {
			return transactionResult, nil
		},
	}
	// get account from  sender address
	acc, _ := client.GetAccount(defaultAddress)
	nonce, err := strconv.ParseUint(acc.Sequence.Nonce, 10, 32)
	require.Nil(t, err)

	assetPb := &ltypes.AssetMessage{
		Amount:           &amount,
		RecipientAddress: recipientAddress,
		Data:             &data,
	}
	asset, err := proto.Marshal(assetPb)
	pubKey := crypto.GetPublicKeyFromSecret(defaultPassphrase)
	privateKey := crypto.GetPrivateKeyFromSecret(defaultPassphrase)

	// init transaction
	txPb := &ltypes.TransactionMessage{
		ModuleID:        &moduleId,
		AssetID:         &assetId,
		Fee:             &fee,
		Asset:           asset,
		Nonce:           &nonce,
		SenderPublicKey: pubKey,
	}
	// marshal transaction data
	txHash, err := proto.Marshal(txPb)
	require.Nil(t, err)

	// sign transaction
	signature, err := sign(txHash, privateKey)
	require.Nil(t, err)
	txPb.Signatures = [][]byte{signature}

	// marshal transaction with Signatures
	txHash, err = proto.Marshal(txPb)
	require.Nil(t, err)

	tx := types.DispatchedTxRequest{Chain: "lisk-testnet", TxHash: hex.EncodeToString(txHash)}
	dispatcher := lisk.NewDispatcher("lisk-testnet", client)
	dpResult := dispatcher.Dispatch(&tx)
	require.Equal(t, dpResult.Success, true)
}

func sign(txBytes []byte, privateKey []byte) ([]byte, error) {
	dst := new(bytes.Buffer)
	//First byte is the network info
	network := networks["testnet"]
	networkBytes, _ := hex.DecodeString(network)
	binary.Write(dst, binary.LittleEndian, networkBytes)

	// Append the transaction ModuleID
	binary.Write(dst, binary.LittleEndian, txBytes)

	return crypto.SignMessage(dst.Bytes(), privateKey), nil
}

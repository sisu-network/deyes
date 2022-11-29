package solana

import (
	"context"
	"crypto/ed25519"
	"os"
	"testing"

	"github.com/cosmos/go-bip39"
	"github.com/gagliardetto/solana-go"
	solanago "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sisu-network/deyes/types"
)

func GetSolanaPrivateKey(mnemonic string) solanago.PrivateKey {
	seed := bip39.NewSeed(mnemonic, "")[:32]
	key := ed25519.NewKeyFromSeed(seed)
	privKey := solanago.PrivateKey(key)

	return privKey
}

func TestDispatch(t *testing.T) {
	mnemonic := os.Getenv("MNEMONIC")
	owner := GetSolanaPrivateKey(mnemonic)

	client := rpc.New(rpc.LocalNet_RPC)

	result, err := client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(
				1*100_000_000,
				owner.PublicKey(),
				solanago.MustPublicKeyFromBase58("CvocQ9ivbdz5rUnTh6zBgxaiR4asMNbXRrG2VPUYpoau"),
			).Build(),
		},
		result.Value.Blockhash,
		solana.TransactionPayer(owner.PublicKey()),
	)
	if err != nil {
		panic(err)
	}

	tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if owner.PublicKey().Equals(key) {
				return &owner
			}
			return nil
		},
	)
	bz, err := tx.MarshalBinary()
	if err != nil {
		panic(err)
	}

	dispatcher := NewDispatcher([]string{rpc.LocalNet_RPC}, []string{rpc.LocalNet_WS})
	dispatcher.Dispatch(&types.DispatchedTxRequest{
		Chain: "solana-devnet",
		Tx:    bz,
	})
}

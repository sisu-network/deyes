package main

import (
	"context"
	"fmt"
	"os"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	cgblockfrost "github.com/echovl/cardano-go/blockfrost"
	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/wallet"
)

const (
	Mnemonic = "art forum devote street sure rather head chuckle guard poverty release quote oak craft enemy"
)

func getWallet() *wallet.Wallet {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	node := cgblockfrost.NewNode(cardano.Testnet, projectId)
	opts := &wallet.Options{Node: node}
	client := wallet.NewClient(opts)

	w, err := client.RestoreWallet("TestWallet", "pass", Mnemonic)
	if err != nil {
		panic(err)
	}

	return w
}

func getAddress() cardano.Address {
	w := getWallet()

	addrs, err := w.Addresses()
	if err != nil {
		panic(err)
	}

	fmt.Println("addrs length = ", len(addrs))
	for _, addr := range addrs {
		fmt.Println(addr)
	}

	return addrs[0]
}

func getTx() *cardano.Tx {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	receiver, err := cardano.NewAddress("receiver_addr")
	sk, err := crypto.NewPrvKey("addr_skey")
	if err != nil {
		panic(err)
	}

	txInput, err := cardano.NewTxInput("txhash", 0, cardano.Coin(2000000))
	if err != nil {
		panic(err)
	}
	txOut, err := cardano.NewTxOutput(receiver, cardano.Coin(1300000))
	if err != nil {
		panic(err)
	}

	txBuilder.AddInputs(txInput)
	txBuilder.AddOutputs(txOut)
	txBuilder.SetTTL(100000)
	txBuilder.SetFee(cardano.Coin(160000))

	tx, err := txBuilder.Build()
	if err != nil {
		panic(err)
	}

	sk.Sign(tx.Bytes())

	fmt.Println(tx.Hex())

	return tx
}

func transfer(addrString string) {
	w := getWallet()
	recipient, err := cardano.NewAddress(addrString)
	if err != nil {
		panic(err)
	}

	balance, err := w.Balance()
	if err != nil {
		panic(err)
	}

	fmt.Println("balance = ", balance)

	txhash, err := w.Transfer(recipient, cardano.Coin(1000000))
	if err != nil {
		panic(err)
	}

	fmt.Println("txhash = ", txhash)
}

func submitTx(tx *cardano.Tx) {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	node := cgblockfrost.NewNode(cardano.Mainnet, "project-id")
	txHash, err := node.SubmitTx(tx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Tx is submitted successfully. hash = ", txHash)
}

func queryBalance() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	api := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: projectId, // Exclude to load from env:BLOCKFROST_PROJECT_ID
			Server:    "https://cardano-testnet.blockfrost.io/api/v0",
		},
	)

	w := getWallet()
	walletAddrs, err := w.Addresses()
	addr := walletAddrs[0]
	fmt.Println("addr.Bech32() = ", addr.Bech32())

	addrs, err := api.AddressUTXOs(context.Background(), addr.Bech32(), blockfrost.APIQueryParams{})
	if err != nil {
		panic(err)
	}

	fmt.Println("utxo txs = ", addrs)

	eparams, err := api.LatestEpochParameters(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("eparams = ", eparams.MinFeeA, eparams.MinFeeB, eparams.MinUtxo)
}

func main() {
	// queryBalance()

	transfer("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
}

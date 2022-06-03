package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	cgblockfrost "github.com/echovl/cardano-go/blockfrost"
	"github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/wallet"
	adacore "github.com/sisu-network/deyes/chains/cardano/core"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"
)

const (
	// wallet address: addr_test1vrfcqffcl8h6j45ndq658qdwdxy2nhpqewv5dlxlmaatducz6k63t
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

func getTx() *cardano.Tx {
	txBuilder := cardano.NewTxBuilder(&cardano.ProtocolParams{})

	receiver, err := cardano.NewAddress("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
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
	tx.Hash()

	txHash, err := node.SubmitTx(tx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Tx is submitted successfully. hash = ", txHash)
}

func query() {
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
	addr := walletAddrs[0].Bech32()
	fmt.Println("addr.Bech32() = ", addr)

	addrUtxos, err := api.AddressUTXOs(context.Background(), addr, blockfrost.APIQueryParams{})
	if err != nil {
		panic(err)
	}

	fmt.Println("utxo txs = ", addrUtxos)

	eparams, err := api.LatestEpochParameters(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("eparams = ", eparams.MinFeeA, eparams.MinFeeB, eparams.MinUtxo)

	addrDetails, err := api.AddressDetails(context.Background(), "addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
	if err != nil {
		panic(err)
	}

	fmt.Println("addrDetails = ", addrDetails.TxCount)
}

func testWatcher() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	chainCfg := config.Chain{
		Chain:      "cardano-testnet",
		BlockTime:  10 * 1000,
		AdjustTime: 1000,
		Rpcs:       []string{"https://cardano-testnet.blockfrost.io/api/v0"},
		RpcSecret:  projectId,
	}

	cfg := config.Deyes{
		PricePollFrequency: 1,
		PriceOracleUrl:     "http://example.com",
		PriceTokenList:     []string{"INVALID_TOKEN"},

		DbHost:   "127.0.0.1",
		DbSchema: "deyes",
		InMemory: true,

		Chains: map[string]config.Chain{"cardano-testnet": chainCfg},
	}

	dbInstance := database.NewDb(&cfg)
	err := dbInstance.Init()
	if err != nil {
		panic(err)
	}

	txsCh := make(chan *types.Txs)
	watcher := adacore.NewWatcher(chainCfg, dbInstance, txsCh)
	watcher.Start()
	watcher.AddWatchAddr("addr_test1vrfcqffcl8h6j45ndq658qdwdxy2nhpqewv5dlxlmaatducz6k63t")

	for {
		select {
		case txs := <-txsCh:
			for _, tx := range txs.Arr {
				fmt.Println("Tx hash = ", tx.Hash)

				txUtxos := new(blockfrost.TransactionUTXOs)
				err := json.Unmarshal(tx.Serialized, txUtxos)
				if err != nil {
					log.Error(err)
					continue
				}

				for _, input := range txUtxos.Inputs {
					fmt.Println(input.Address, input.Amount, input.TxHash)
				}
				fmt.Println()

				fmt.Println("==========")
				for _, output := range txUtxos.Outputs {
					fmt.Println(output.Address, output.Amount)
				}
			}
		}
	}
}

func testTxMetadata() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	api := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: projectId, // Exclude to load from env:BLOCKFROST_PROJECT_ID
			Server:    "https://cardano-mainnet.blockfrost.io/api/v0",
		},
	)

	metaArr, err := api.TransactionMetadata(context.Background(), "191d2579d12394672b8a55fd2f57e48036d0a8650863fdb96ae6cd37ae1caf66")
	if err != nil {
		panic(err)
	}

	for _, meta := range metaArr {
		fmt.Println(meta.JsonMetadata)
	}
}

func main() {
	// query()
	// transfer("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
	// testWatcher()

	testTxMetadata()
}

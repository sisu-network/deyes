package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
	"github.com/echovl/cardano-go"
	cgblockfrost "github.com/echovl/cardano-go/blockfrost"
	"github.com/echovl/cardano-go/wallet"
	adacore "github.com/sisu-network/deyes/chains/cardano/core"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"

	cardanobf "github.com/echovl/cardano-go/blockfrost"
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

func getApi() blockfrost.APIClient {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	api := blockfrost.NewAPIClient(
		blockfrost.APIClientOptions{
			ProjectID: projectId,
			Server:    "https://cardano-testnet.blockfrost.io/api/v0",
		},
	)
	return api
}

func getCardanoNode() cardano.Node {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	node := cardanobf.NewNode(cardano.Testnet, projectId)
	return node
}

func transfer(addrString string) {
	w := getWallet()

	recipient, err := cardano.NewAddress(addrString)
	if err != nil {
		panic(err)
	}

	txhash, err := w.Transfer(recipient, cardano.NewValue(cardano.Coin(1000000)))
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
	api := getApi()
	w := getWallet()

	walletAddrs, err := w.Addresses()
	addr := walletAddrs[0].Bech32()
	fmt.Println("addr.Bech32() = ", addr)
	addrUtxos, err := api.AddressUTXOs(context.Background(), addr, blockfrost.APIQueryParams{})
	if err != nil {
		panic(err)
	}
	fmt.Println("utxo txs = ", addrUtxos)
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

func getProtocolParams(bfParams blockfrost.EpochParameters) *cardano.ProtocolParams {
	keyDeposit, err := strconv.Atoi(bfParams.KeyDeposit)
	if err != nil {
		panic(err)
	}

	minUtxo, err := strconv.Atoi(bfParams.MinUtxo)

	return &cardano.ProtocolParams{
		MinFeeA:          cardano.Coin(bfParams.MinFeeA),
		MinFeeB:          cardano.Coin(bfParams.MinFeeB),
		KeyDeposit:       cardano.Coin(keyDeposit),
		CoinsPerUTXOWord: cardano.Coin(minUtxo),
	}
}

func testTxHash() {
	api := getApi()
	bfParams, err := api.LatestEpochParameters(context.Background())
	if err != nil {
		panic(err)
	}

	w := getWallet()
	addrs, err := w.Addresses()
	if err != nil {
		panic(err)
	}
	senderAddr := addrs[0]
	fmt.Println(senderAddr)

	utxos, err := api.AddressUTXOs(context.Background(), senderAddr.String(), blockfrost.APIQueryParams{})
	if err != nil {
		panic(err)
	}
	fmt.Println("utxos = ", utxos)

	protocolParams := getProtocolParams(bfParams)
	txBuilder := cardano.NewTxBuilder(protocolParams)

	receiver, err := cardano.NewAddress("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")

	txHash, err := cardano.NewHash32("bc82779c18b98f0f5628b0cae12af618020e5388258d3bcce936c380583298dc")
	if err != nil {
		panic(err)
	}

	txInput := cardano.NewTxInput(txHash, 0, cardano.NewValue(994171615))
	txOut := cardano.NewTxOutput(receiver, cardano.NewValue(1000000))

	txBuilder.AddInputs(txInput)
	txBuilder.AddOutputs(txOut)
	txBuilder.SetFee(cardano.Coin(1000000))

	block, err := api.BlockLatest(context.Background())
	if err != nil {
		panic(err)
	}

	txBuilder.SetTTL(uint64(block.Slot) + 1200)
	// Sign transaction
	key, _ := w.Keys()
	txBuilder.Sign(key)

	txBuilder.AddChangeIfNeeded(senderAddr)

	tx, err := txBuilder.Build()
	if err != nil {
		panic(err)
	}

	localHash, err := tx.Hash()
	if err != nil {
		panic(err)
	}
	fmt.Println("localHash = ", localHash)

	node := getCardanoNode()
	hash, err := node.SubmitTx(tx)
	if err != nil {
		panic(err)
	}

	fmt.Println("hash = ", hash)
}

func queryTxUtxo() {
	api := getApi()

	txUtxos, err := api.TransactionUTXOs(context.Background(), "6c9025b0fe319e1015665973e1d8bfc03d8dc7de0d211f82fb863df8b175a4aa")
	if err != nil {
		panic(err)
	}

	for _, input := range txUtxos.Inputs {
		for _, amount := range input.Amount {
			fmt.Println("Amount: ", amount.Quantity, amount.Unit)
		}
	}

	fmt.Println("=============")

	asset, err := api.Asset(context.Background(), "6b8d07d69639e9413dd637a1a815a7323c69c86abbafb66dbfdb1aa7")
	if err != nil {
		panic(err)
	}

	fmt.Printf("asset = %s, policy = %s, name = %s, fingerprint = %s\n", asset.Asset, asset.PolicyId, asset.AssetName, asset.Fingerprint)
	assets, err := api.AssetsByPolicy(context.Background(), "6b8d07d69639e9413dd637a1a815a7323c69c86abbafb66dbfdb1aa7")
	if err != nil {
		panic(err)
	}

	for _, asset := range assets {
		fmt.Printf("asset = %s, quantity = %s, metadata = %v\n", asset.Asset, asset.Quantity, asset.Metadata)
		decode, err := hex.DecodeString(asset.Asset)
		if err != nil {
			fmt.Println("err = ", err)
		} else {
			fmt.Println("Decode = ", len(decode), decode)
		}
	}
}

func queryAddressTransaction() {
	api := getApi()
	txs, err := api.AddressTransactions(context.Background(), "addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy", blockfrost.APIQueryParams{
		From: "3604437",
		To:   "3604437",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("txs = ", txs)
}

func main() {
	// query()
	// transfer("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
	// testWatcher()
	queryTxUtxo()
	// testBlockfrostClient()

	// testTxHash()

	// queryAddressTransaction()
}

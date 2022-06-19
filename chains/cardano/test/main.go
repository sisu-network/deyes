package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/blockfrost/blockfrost-go"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	"github.com/echovl/cardano-go"
	cgblockfrost "github.com/echovl/cardano-go/blockfrost"
	"github.com/echovl/cardano-go/wallet"
	carcore "github.com/sisu-network/deyes/chains/cardano/core"
	"github.com/sisu-network/deyes/chains/cardano/utils"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"

	cardanobf "github.com/echovl/cardano-go/blockfrost"
)

// Miscellaneous test for cardano watcher
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

func transfer(addr string, value int) {
	w := getWallet()
	receiver, err := cardano.NewAddress(addr)
	if err != nil {
		panic(err)
	}

	hash, err := w.Transfer(receiver, cardano.NewValue(cardano.Coin(value)), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash = ", hash.String())
}

func testWatcher() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	chainCfg := config.Chain{
		Chain:      "cardano-testnet",
		BlockTime:  20 * 1000,
		AdjustTime: 2000,
		Rpcs:       []string{"https://cardano-testnet.blockfrost.io/api/v0"},
		RpcSecret:  projectId,
	}

	cfg := config.Deyes{
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
	watcher := carcore.NewWatcher(chainCfg, dbInstance, txsCh, carcore.NewBlockfrostClient(blockfrost.APIClientOptions{
		ProjectID: projectId,
		Server:    chainCfg.Rpcs[0],
	}))
	watcher.Start()
	watcher.AddWatchAddr("addr_test1vrfcqffcl8h6j45ndq658qdwdxy2nhpqewv5dlxlmaatducz6k63t")

	go func() {
		w := getWallet()
		receiver, err := cardano.NewAddress("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
		if err != nil {
			panic(err)
		}

		hash, err := w.Transfer(receiver, cardano.NewValue(1000000), nil)
		if err != nil {
			panic(err)
		}
		fmt.Println("Hash = ", hash.String())
	}()

	log.Info("listening to new txs...")
	for {
		select {
		case txs := <-txsCh:
			for _, tx := range txs.Arr {
				log.Info("Tx hash = ", tx.Hash)

				txUtxos := new(blockfrost.TransactionUTXOs)
				err := json.Unmarshal(tx.Serialized, txUtxos)
				if err != nil {
					log.Error(err)
					continue
				}

				for _, input := range txUtxos.Inputs {
					log.Info(input.Address, input.Amount, input.TxHash)
				}
				log.Info()

				log.Info("==========")
				for _, output := range txUtxos.Outputs {
					log.Info(output.Address, output.Amount)
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

func constructTx(api blockfrost.APIClient, senderAddr cardano.Address) *cardano.TxBuilder {
	bfParams, err := api.LatestEpochParameters(context.Background())
	if err != nil {
		panic(err)
	}

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

	return txBuilder
}

func getTestTx(w *wallet.Wallet) *cardano.Tx {
	api := getApi()
	if w == nil {
		w = getWallet()
	}
	addrs, err := w.Addresses()
	if err != nil {
		panic(err)
	}
	senderAddr := addrs[0]
	fmt.Println(senderAddr)

	txBuilder := constructTx(api, senderAddr)
	// Sign transaction
	key, _ := w.Keys()
	_ = key
	// txBuilder.Sign(key)

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

	return tx
}

func testTxHash() {
	tx := getTestTx(nil)
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

func testBlockfrostClient() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	client := carcore.NewBlockfrostClient(
		blockfrost.APIClientOptions{
			ProjectID: projectId,
			Server:    "https://cardano-testnet.blockfrost.io/api/v0",
		},
	)

	txsIn, err := client.NewTxs(3637884, map[string]bool{"addr_test1qqdnqmpjwac5e8j8gf7gsa75p99rf07rsc63fju0w5kywj20aczxffwdmqewegjqzc24074fk6tqgydujpez0aslcd7srp9cvt": true})
	if err != nil {
		panic(err)
	}

	for _, txIn := range txsIn {
		log.Infof("TxIn item = %+v\n", txIn)
		log.Infof("additional info: %+v\n", txIn.TxAdditionInfo)
	}
}

func transferWithMetadata(destChain, destToken, destRecipient, cardanoGwAddr string, value uint64) {
	w := getWallet()
	receiver, err := cardano.NewAddress(cardanoGwAddr)
	if err != nil {
		panic(err)
	}

	metadata := cardano.Metadata{
		0: map[string]interface{}{
			"destination_chain":         destChain,
			"destination_recipient":     destRecipient,
			"destination_token_address": destToken,
		},
	}

	hash, err := w.Transfer(receiver, cardano.NewValue(cardano.Coin(value)), metadata)
	if err != nil {
		panic(err)
	}
	fmt.Println("Hash = ", hash.String())
}

func getAddressFromBytes(bz []byte) cardano.Address {
	keyHash, err := cardano.Blake224Hash(bz)
	if err != nil {
		panic(err)
	}

	payment := cardano.StakeCredential{Type: cardano.KeyCredential, KeyHash: keyHash}
	addr, err := cardano.NewEnterpriseAddress(cardano.Testnet, payment)
	if err != nil {
		panic(err)
	}

	fmt.Println("addr = ", addr)
	return addr
}

func randByteArray(n int, seed int) []byte {
	rand.Seed(int64(seed))

	bz := make([]byte, n)
	for i := 0; i < n; i++ {
		bz[i] = byte(rand.Intn(256))
	}

	return bz
}

func testSigning() {
	// api := getApi()
	seed := randByteArray(32, 98)
	edwardsPrivate, edwardsPublic := edwards.PrivKeyFromSecret(seed)
	bz := edwardsPublic.Serialize()

	if pubkey, err := edwards.ParsePubKey(bz); err == nil {
		if pubkey.X.Cmp(edwardsPublic.X) != 0 || pubkey.Y.Cmp(edwardsPublic.Y) != 0 {
			panic("Key not equal")
		}
	} else {
		panic(err)
	}

	node := getCardanoNode()
	sender := utils.GetAddressFromCardanoPubkey(bz)
	log.Info("Sender = ", sender)
	receiver, err := cardano.NewAddress("addr_test1vqxyzpun2fpqafvkxxxceu5r8yh4dccy6xdcynnchd4dr7qtjh44z")
	if err != nil {
		panic(err)
	}

	hash, err := utils.Transfer(node, cardano.Testnet, edwardsPrivate, sender, receiver, cardano.NewValue(1e6))
	if err != nil {
		panic(err)
	}

	log.Info("Transaction hash = ", hash)
}

func main() {
	// query()
	// transfer("addr_test1vqxyzpun2fpqafvkxxxceu5r8yh4dccy6xdcynnchd4dr7qtjh44z", 10_000_000)
	transferWithMetadata("ganache1",
		"0x3A84fBbeFD21D6a5ce79D54d348344EE11EBd45C",
		"0x215375950B138B9f5aDfaEb4dc172E8AD1dDe7f5",
		"addr_test1vq987lkjn3eh5pdj8rhg3qq2m24hhpecleytch2q8mk0nyqdmcvhx",
		30_000_000)
	//testBlockfrostClient()
	//testWatcher()

	//testSigning()
}

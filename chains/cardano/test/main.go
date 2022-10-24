package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/joho/godotenv"
	"github.com/tyler-smith/go-bip39"

	"github.com/BurntSushi/toml"
	"github.com/blockfrost/blockfrost-go"
	"github.com/decred/dcrd/dcrec/edwards/v2"
	"github.com/echovl/cardano-go"
	cgblockfrost "github.com/echovl/cardano-go/blockfrost"
	cardanocrypto "github.com/echovl/cardano-go/crypto"
	"github.com/echovl/cardano-go/wallet"
	chainscardano "github.com/sisu-network/deyes/chains/cardano"
	"github.com/sisu-network/deyes/chains/cardano/utils"
	chainstypes "github.com/sisu-network/deyes/chains/types"
	"github.com/sisu-network/deyes/config"
	"github.com/sisu-network/deyes/database"
	"github.com/sisu-network/deyes/types"
	"github.com/sisu-network/lib/log"

	cardanobf "github.com/echovl/cardano-go/blockfrost"
	providertypes "github.com/sisu-network/deyes/chains/cardano/types"
)

var (
	WrapADA = cardano.NewAssetName("WRAP_ADA")
)

func getWallet() *wallet.Wallet {
	projectId := os.Getenv("BLOCKFROST_SECRET")
	Mnemonic := os.Getenv("MNEMONIC")

	if len(projectId) == 0 {
		panic("project id is empty")
	}

	node := cgblockfrost.NewNode(cardano.Preprod, projectId)
	opts := &wallet.Options{Node: node}
	client := wallet.NewClient(opts)

	w, err := client.RestoreWallet("sisu", "12345678910", Mnemonic)
	addr, err := w.AddAddress()
	if err != nil {
		panic(err)
	}

	log.Info("Address = ", addr.String())

	return w
}

func getKey(name, password, mnemonic string) cardanocrypto.XPrvKey {
	entropy, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}

	// This logic is taken from here:
	// https://github.com/sisu-network/cardano-go/blob/81453a2fe980c23b9af74a632928907a55dfc692/wallet/wallet.go#L183
	purposeIndex := uint32(1852 + 0x80000000)
	coinTypeIndex := uint32(1815 + 0x80000000)
	accountIndex := uint32(0x80000000)
	externalChainIndex := uint32(0x0)

	rootKey := cardanocrypto.NewXPrvKeyFromEntropy(entropy, password)
	accountKey := rootKey.Derive(uint32(purposeIndex)).
		Derive(coinTypeIndex).
		Derive(accountIndex)
	chainKey := accountKey.Derive(externalChainIndex)
	addr0Key := chainKey.Derive(0)

	return addr0Key
}

func loadConfig() *config.Deyes {
	tomlFile := "./deyes.toml"
	if _, err := os.Stat(tomlFile); os.IsNotExist(err) {
		panic(err)
	}

	cfg := new(config.Deyes)
	_, err := toml.DecodeFile(tomlFile, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
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
	log.Info("Hash = ", hash.String())
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
		Rpcs:       []string{"https://cardano-preprod.blockfrost.io/api/v0"},
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

	provider := chainscardano.NewBlockfrostProvider(chainCfg)
	txsCh := make(chan *types.Txs)
	watcher := chainscardano.NewWatcher(chainCfg, dbInstance, txsCh,
		make(chan *chainstypes.TrackUpdate, 3),
		chainscardano.NewDefaultCardanoClient(provider, blockfrost.CardanoTestNet+"/tx/submit", projectId))
	watcher.Start()
	watcher.SetVault("addr_test1vrfcqffcl8h6j45ndq658qdwdxy2nhpqewv5dlxlmaatducz6k63t", "")

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
		log.Info("Hash = ", hash.String())
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

func constructTx(api chainscardano.Provider, senderAddr cardano.Address) *cardano.TxBuilder {
	protocolParams, err := api.LatestEpochParameters(context.Background())
	if err != nil {
		panic(err)
	}

	utxos, err := api.AddressUTXOs(context.Background(), senderAddr.String(), providertypes.APIQueryParams{})
	if err != nil {
		panic(err)
	}

	txBuilder := cardano.NewTxBuilder(protocolParams)

	receiver, err := cardano.NewAddress("addr_test1vqyqp03az6w8xuknzpfup3h7ghjwu26z7xa6gk7l9j7j2gs8zfwcy")
	if err != nil {
		panic(err)
	}

	for _, utxo := range utxos {
		txBuilder.AddInputs(&cardano.TxInput{TxHash: utxo.TxHash, Index: utxo.Index, Amount: utxo.Amount})
	}
	txOut := cardano.NewTxOutput(receiver, cardano.NewValue(1_000_000))

	txBuilder.AddOutputs(txOut)
	txBuilder.AddChangeIfNeeded(senderAddr)
	txBuilder.SetFee(cardano.Coin(200_000))

	block, err := api.BlockLatest(context.Background())
	if err != nil {
		panic(err)
	}

	txBuilder.SetTTL(uint64(block.Slot) + 1200)

	return txBuilder
}

func queryTxUtxo() {
	api := getApi()

	txUtxos, err := api.TransactionUTXOs(context.Background(), "6c9025b0fe319e1015665973e1d8bfc03d8dc7de0d211f82fb863df8b175a4aa")
	if err != nil {
		panic(err)
	}

	for _, input := range txUtxos.Inputs {
		for _, amount := range input.Amount {
			log.Info("Amount: ", amount.Quantity, amount.Unit)
		}
	}

	log.Info("=============")

	asset, err := api.Asset(context.Background(), "6b8d07d69639e9413dd637a1a815a7323c69c86abbafb66dbfdb1aa7")
	if err != nil {
		panic(err)
	}

	log.Verbosef("asset = %s, policy = %s, name = %s, fingerprint = %s\n", asset.Asset, asset.PolicyId, asset.AssetName, asset.Fingerprint)
	assets, err := api.AssetsByPolicy(context.Background(), "6b8d07d69639e9413dd637a1a815a7323c69c86abbafb66dbfdb1aa7")
	if err != nil {
		panic(err)
	}

	for _, asset := range assets {
		log.Verbosef("asset = %s, quantity = %s, metadata = %v\n", asset.Asset, asset.Quantity, asset.Metadata)
		decode, err := hex.DecodeString(asset.Asset)
		if err != nil {
			log.Error("err = ", err)
		} else {
			log.Info("Decode = ", len(decode), decode)
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

	log.Info("txs = ", txs)
}

func testBlockfrostClient() {
	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		panic("project id is empty")
	}

	chainCfg := config.Chain{
		Chain:      "cardano-testnet",
		BlockTime:  20 * 1000,
		AdjustTime: 2000,
		Rpcs:       []string{"https://cardano-preprod.blockfrost.io/api/v0"},
		RpcSecret:  projectId,
	}

	provider := chainscardano.NewBlockfrostProvider(chainCfg)
	client := chainscardano.NewDefaultCardanoClient(
		provider, blockfrost.CardanoTestNet+"/tx/submit", projectId,
	)

	txsIn, err := client.NewTxs(3654812, "addr_test1vpa9x6a7r4cwg6r052yj25usa2gkxarps8zecfmtx4p7erqwtfq45")
	if err != nil {
		panic(err)
	}

	for _, txIn := range txsIn {
		log.Infof("TxIn item = %+v\n", txIn)
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
			"chain":     destChain,
			"recipient": destRecipient,
		},
	}

	hash, err := w.Transfer(receiver, cardano.NewValue(cardano.Coin(value)), metadata)
	if err != nil {
		panic(err)
	}

	log.Info("Hash = ", hash.String())
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

	log.Info("addr = ", addr)
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

func queryBalance() {
	node := getCardanoNode()
	addr, err := cardano.NewAddress("addr_test1vzxv7v8r5v3umgu9d3v3968sl603s2jkdqk58u6c2v9zmdqqdyvd7")
	if err != nil {
		panic(err)
	}

	value, err := utils.Balance(node, addr)
	if err != nil {
		panic(err)
	}

	log.Info("Balance = ", value)
}

// Get WRAP_ADA (hex: 575241505f414441) token.
// PolicyID (dc89700b3adf88f6b520aba2f3cfa4c26fa7a19bd8eadf430d73b9d4) got from there:
// https://explorer.cardano-testnet.iohkdev.io/en/transaction?id=31c019b737edc7b54ae60a87f372f860715e8bb02b979ed853395ccbf4bf0209
func getMultiAsset(amt uint64) *cardano.MultiAsset {
	policyHash, err := cardano.NewHash28("dc89700b3adf88f6b520aba2f3cfa4c26fa7a19bd8eadf430d73b9d4")
	if err != nil {
		err := fmt.Errorf("error when parsing policyID hash: %v", err)
		panic(err)
	}

	policyID := cardano.NewPolicyIDFromHash(policyHash)
	log.Info("policyID = ", policyID.String())

	asset := cardano.NewAssets().Set(WrapADA, cardano.BigNum(amt*1_000_000))
	return cardano.NewMultiAsset().Set(policyID, asset)
}

func transferMultiAsset(recipient string, amount uint64) {
	recipientAddr, err := cardano.NewAddress(recipient)
	if err != nil {
		panic(err)
	}
	w := getWallet()
	addr, err := w.AddAddress()
	if err != nil {
		panic(err)
	}
	log.Verbose("Sender address = ", addr.String())

	metadata := cardano.Metadata{
		0: map[string]interface{}{
			"chain":      "ganache1",
			"recipient":  "0x215375950B138B9f5aDfaEb4dc172E8AD1dDe7f5",
			"native_ada": 1,
		},
	}

	txHash, err := w.Transfer(recipientAddr, cardano.NewValueWithAssets(cardano.Coin(amount), getMultiAsset(1e3)), metadata)
	if err != nil {
		panic(err)
	}

	log.Info("txHash = ", txHash)
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func testUtxos() {
	api := getApi()
	txHashes, err := api.TransactionUTXOs(context.Background(), "c3998f845e159598f566fe1418d86b22296953a83bda2a96eab411f9ff05a0c2")
	if err != nil {
		panic(err)
	}

	log.Info(txHashes.Inputs[0].Address)
}

func testWatcherSyncDB() {
	cfg := loadConfig()
	log.Verbose(cfg)

	db, err := chainscardano.ConnectDB(cfg.Chains["cardano-testnet"].SyncDB)
	if err != nil {
		panic(err)
	}

	syncDB := chainscardano.NewSyncDBConnector(db)

	cardanoClient := chainscardano.NewDefaultCardanoClient(syncDB, "", "")
	txs, err := cardanoClient.NewTxs(229733, "addr_test1qrfut328td5krhgjk7eh6k9uh5ek6egk2xf0wxqxjaqq2jrtw0ge2rnesr84874t3cz5gft6y9rck5qzlmdf0ygtredsddxavz")
	if err != nil {
		panic(err)
	}

	if len(txs) == 0 {
		panic("Transaction list is empty")
	}

	for _, tx := range txs {
		log.Verbosef("tx = %s\n", tx.Hash)
	}
}

func testDbSyncSubmitTx() {
	cfg := loadConfig()
	log.Verbose(cfg)

	db, err := chainscardano.ConnectDB(cfg.Chains["cardano-testnet"].SyncDB)
	if err != nil {
		panic(err)
	}
	syncDB := chainscardano.NewSyncDBConnector(db)
	sender, err := cardano.NewAddress("addr_test1vzdw29a37lf0c3xv0rsdksreapcah4y4kucae35ujz5cucs69z3v6")
	if err != nil {
		panic(err)
	}

	txBuilder := constructTx(syncDB, sender)
	client := chainscardano.NewDefaultCardanoClient(syncDB, cfg.Chains["cardano-testnet"].SyncDB.SubmitURL, "")

	txBuilder.Sign(getKey(os.Getenv("USER"), os.Getenv("PASSWORD"), os.Getenv("MNEMONIC")).PrvKey())

	tx, err := txBuilder.Build()
	if err != nil {
		panic(err)
	}

	hash, err := client.SubmitTx(tx)
	if err != nil {
		panic(err)
	}

	log.Verbose("Hash = ", hash)
}

func testSyncDb() {
	cfg := loadConfig()
	log.Verbose(cfg)

	db, err := chainscardano.ConnectDB(cfg.Chains["cardano-testnet"].SyncDB)
	if err != nil {
		panic(err)
	}
	syncDB := chainscardano.NewSyncDBConnector(db)

	block, err := syncDB.Block(context.Background(), "12345")
	if err != nil {
		panic(err)
	}

	log.Verbose("block = ", block)
}

func main() {
	loadEnv()

	testSyncDb()

	// testWatcherSyncDB()
	// testDbSyncSubmitTx()
	// transfer("addr_test1vqp9rec4rzljt64ykvp5qersc8aldhhr94uae9p9gpmr88s4494xt", 150*1_000_000)

	// transferWithMetadata("ganache1",
	// 	"0x3A84fBbeFD21D6a5ce79D54d348344EE11EBd45C",
	// 	"0x215375950B138B9f5aDfaEb4dc172E8AD1dDe7f5",
	// 	"addr_test1vrwxrgqf9fplssrkc27k2zt0rm6d8as4v3q3zu34annh9dg4hnp4t",
	// 	1_000_000)
}

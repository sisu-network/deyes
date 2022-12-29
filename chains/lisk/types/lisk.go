package types

type Transaction struct {
	Id              string   `json:"id"`
	ModuleAssetId   string   `json:"moduleAssetId"`
	ModuleAssetName string   `json:"moduleAssetName"`
	Height          int64    `json:"height"`
	Nonce           string   `json:"nonce"`
	Block           string   `json:"block"`
	Sender          Sender   `json:"sender"`
	Signatures      []string `json:"signatures"`
	Asset           Asset    `json:"asset"`
	IsPending       bool     `json:"isPending"`
}

type TransactionBlock struct {
	Id        string `json:"id"`
	Height    int64  `json:"height"`
	Timestamp int64  `json:"timestamp"`
}

type Sender struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Username  string `json:"username"`
}

type Asset struct {
	Amount    string         `json:"amount"`
	Data      string         `json:"data"`
	Recipient AssetRecipient `json:"recipient"`
}

type AssetRecipient struct {
	Address string `json:"address"`
}

type ResponseWrapper struct {
	Data []map[string]interface{} `json:"data"`
	Meta Meta
}

type Meta struct {
	Count  int64 `json:"count"`
	Offset int   `json:"offset"`
	Total  int   `json:"total"`
}

type Block struct {
	Id                        string        `json:"id"`
	Height                    uint64        `json:"height"`
	Version                   int           `json:"version"`
	Timestamp                 int64         `json:"timestamp"`
	GeneratorAddress          string        `json:"generatorAddress"`
	GeneratorPublicKey        string        `json:"generatorPublicKey"`
	GeneratorUsername         string        `json:"generatorUsername"`
	TransactionRoot           string        `json:"transactionRoot"`
	Signature                 string        `json:"signature"`
	PreviousBlockId           string        `json:"previousBlockId"`
	NumberOfTransactions      int64         `json:"numberOfTransactions"`
	TotalForged               string        `json:"totalForged"`
	TotalBurnt                string        `json:"totalBurnt"`
	TotalFee                  string        `json:"totalFee"`
	Reward                    string        `json:"reward"`
	IsFinal                   bool          `json:"isFinal"`
	MaxHeightPreviouslyForged int64         `json:"maxHeightPreviouslyForged"`
	MaxHeightPrevoted         int64         `json:"maxHeightPrevoted"`
	SeedReveal                string        `json:"seedReveal"`
	Transactions              []Transaction `json:"transactions"`
}

type Params struct {
	Sort    string `json:"sort"`
	Height  string `json:"height"`
	BlockId string `json:"blockId"`
}

package types

type GetBlockRequest struct {
	TransactionDetails             string `json:"transactionDetails"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
}

type Instruction struct {
	ProgramIdIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}

type TransactionMeta struct {
	Fee uint64      `json:"fee"`
	Err interface{} `json:"err"`
}

type TransactionMessage struct {
	AccountKeys  []string      `json:"accountKeys"`
	Instructions []Instruction `json:"instructions"`
}

type TransactionInner struct {
	Signatures []string            `json:"signatures"`
	Message    *TransactionMessage `json:"Message"`
}

type Transaction struct {
	Meta             *TransactionMeta  `json:"meta"`
	TransactionInner *TransactionInner `json:"transaction"`
}

type Block struct {
	BlockHeight  int            `json:"blockHeight"`
	Transactions []*Transaction `json:"transactions"`
	ParentSlot   int            `json:"parentSlot"`
}

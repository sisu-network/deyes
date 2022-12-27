package lisk

type Payload struct {
	Jsonrpc string
	Id      string
	Method  string
	Params  TransactionResponse
}

type TransactionResponse struct {
	Transaction string
}

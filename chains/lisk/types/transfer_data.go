package types

type TransferData struct {
	ChainId   uint64
	Recipient []byte
	Token     string
	Amount    uint64
}

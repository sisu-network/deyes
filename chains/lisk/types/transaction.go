package types

import "fmt"

type Transaction struct {
	Id              string            `json:"id"`
	ModuleAssetId   string            `json:"moduleAssetId"`
	ModuleAssetName string            `json:"moduleAssetName"`
	Height          int64             `json:"height"`
	Nonce           string            `json:"nonce"`
	Block           *TransactionBlock `json:"block"`
	Sender          *Sender           `json:"sender"`
	Signatures      []string          `json:"signatures"`
	Asset           *Asset            `json:"asset"`
	IsPending       bool              `json:"isPending"`
}

func (tx *Transaction) Validate() error {
	if tx.Asset == nil {
		return fmt.Errorf("Tx asset is nil")
	}

	if tx.Asset.Recipient == nil {
		return fmt.Errorf("Asset recipient is nil")
	}

	if tx.Sender == nil {
		return fmt.Errorf("Tx sender is nil")
	}

	return nil
}

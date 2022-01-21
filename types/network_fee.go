package types

type GasPriceRequest struct {
	Chain    string `json:"chain,omitempty"`
	Height   int64  `json:"height,omitempty"`
	GasPrice int64  `json:"gas_price,omitempty"`
}

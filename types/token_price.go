package types

type TokenPrices []*TokenPrice

type TokenPrice struct {
	Id       string
	PublicId string
	Price    float32
}

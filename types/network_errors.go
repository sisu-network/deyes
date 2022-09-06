package types

type DispatchError int

const (
	ErrNil DispatchError = iota // no error
	ErrGeneric
	ErrNotEnoughBalance
	ErrMarshal
	ErrSubmitTx
)

package chains

import (
	"github.com/sisu-network/deyes/types"
)

type Dispatcher interface {
	Start()
	Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult
}

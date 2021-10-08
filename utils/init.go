package utils

import (
	etypes "github.com/ethereum/go-ethereum/core/types"
)

var (
	ethSigners map[string]etypes.Signer
)

func init() {
	ethSigners = make(map[string]etypes.Signer)

	// TODO: Add correct signer for each chain. For now, use NewEIP2930Signer for all chains.
	ethSigners["eth"] = etypes.NewEIP2930Signer(GetChainIntFromId("eth"))
	ethSigners["sisu-eth"] = etypes.NewEIP2930Signer(GetChainIntFromId("sisu-eth"))
}

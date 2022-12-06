package solana

import (
	"context"
	"fmt"
	"time"

	solanatypes "github.com/sisu-network/deyes/chains/solana/types"
	"github.com/sisu-network/lib/log"
	"github.com/ybbus/jsonrpc/v3"
)

type BlockResult struct {
	Skip  bool
	Slot  uint64
	Block *solanatypes.Block
}

type fetcher struct {
	n            uint64
	startingSlot uint64
	clients      []jsonrpc.RPCClient
	blockCh      chan *BlockResult
}

func newFetcher(clients []jsonrpc.RPCClient, n, startingSlot uint64, blockCh chan *BlockResult) *fetcher {
	return &fetcher{
		n:            n,
		startingSlot: startingSlot,
		blockCh:      blockCh,
		clients:      clients,
	}
}

func (f *fetcher) start() {
	slot := f.startingSlot

	errCount := 0

	for {
		time.Sleep(time.Second)

		block, err := f.getBlockNumber(slot)

		if err != nil {
			rpcErr, ok := err.(*jsonrpc.RPCError)
			if ok {
				// -32007: Slot 171913340 was skipped, or missing due to ledger jump to recent snapshot
				// -32015: Transaction version (0) is not supported by the requesting client. Please try the request again with the following configuration parameter: \"maxSupportedTransactionVersion\": 0"
				if rpcErr.Code == -32007 || rpcErr.Code == -32015 {
					f.blockCh <- &BlockResult{Skip: true, Slot: slot}
					// Slot is skipped, try the next one.
					errCount = 0
					slot += f.n
				}
			} else {
				log.Warn("Solana fetcher error: err = ", err)
				errCount++
				if errCount == 10 {
					// We reach the maximum error count for unknown reason. Skip this blog
					log.Error("Max retry reached. Skip this slot ", slot)
					f.blockCh <- &BlockResult{Skip: true, Slot: slot}
					errCount = 0
					slot += f.n
				}
			}
		} else {
			if block != nil {
				if block.Transactions == nil {
					log.Error("Err is not nil but transactions list is nil. block = ", block)
					f.blockCh <- &BlockResult{Skip: true, Slot: slot}
				} else {
					f.blockCh <- &BlockResult{Skip: false, Slot: slot, Block: block}
				}
			}

			slot += f.n
			errCount = 0
		}
	}
}

func (f *fetcher) getBlockNumber(slot uint64) (*solanatypes.Block, error) {
	return executeWithClients(f.clients, func(client jsonrpc.RPCClient) (*solanatypes.Block, bool, error) {
		var request = &solanatypes.GetBlockRequest{
			TransactionDetails:             "full",
			MaxSupportedTransactionVersion: 100,
		}

		res, err := client.Call(context.Background(), "getBlock", slot, request)
		if err != nil {
			return nil, false, err
		}

		if res.Error != nil {
			return nil, true, res.Error
		}

		block := new(solanatypes.Block)
		err = res.GetObject(&block)
		if err != nil {
			return block, true, err
		}

		if block == nil {
			err := fmt.Errorf("Error is nil but block is also nil. Slot = %d", slot)
			log.Error(err)

			return block, true, err
		}

		return block, true, nil
	})
}

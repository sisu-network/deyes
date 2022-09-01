package core

import (
	"context"
	"time"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sisu-network/lib/log"
)

const (
	MaxReceiptRetry = 5
)

// txReceiptRequest is a data structure for this watcher to send request to the receipt fetcher.
type txReceiptRequest struct {
	blockNumber int64
	blockHash   string
	txs         []*etypes.Transaction
}

// txReceiptResponse is a data structure for the receipt fetcher to return its result
type txReceiptResponse struct {
	blockNumber int64
	blockHash   string
	txs         []*etypes.Transaction
	receipts    []*etypes.Receipt
}

type receiptFetcher interface {
	start()
	fetchReceipts(block int64, txs []*etypes.Transaction)
}

type defaultReceiptFetcher struct {
	chain      string
	requestCh  chan *txReceiptRequest
	responseCh chan *txReceiptResponse

	clients []EthClient
	txQueue []*etypes.Transaction
}

func newReceiptFetcher(responseCh chan *txReceiptResponse, clients []EthClient, chain string) receiptFetcher {
	return &defaultReceiptFetcher{
		chain:      chain,
		requestCh:  make(chan *txReceiptRequest, 20),
		responseCh: responseCh,
		clients:    clients,
		txQueue:    make([]*etypes.Transaction, 0),
	}
}

func (rf *defaultReceiptFetcher) start() {
	for {
		request := <-rf.requestCh

		retry := 0
		response := &txReceiptResponse{
			blockNumber: request.blockNumber,
			blockHash:   request.blockHash,
			txs:         make([]*etypes.Transaction, 0),
			receipts:    make([]*etypes.Receipt, 0),
		}

		for {
			rf.txQueue = append(rf.txQueue, request.txs...)
			tx := rf.txQueue[0]
			ok := false
			for _, client := range rf.clients {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				receipt, err := client.TransactionReceipt(ctx, tx.Hash())
				cancel()

				if err == nil && receipt != nil {
					ok = true
					response.txs = append(response.txs, tx)
					response.receipts = append(response.receipts, receipt)
					break
				}
			}

			if ok {
				retry = 0
				rf.txQueue = rf.txQueue[1:]
				if len(rf.txQueue) == 0 {
					break
				}
			} else {
				if retry == MaxReceiptRetry {
					log.Errorf("cannot get receipt for tx with hash %s on chain %s", tx.Hash().String(), rf.chain)
					rf.txQueue = rf.txQueue[1:]
					continue
				}

				retry++
				time.Sleep(time.Second * 5)
			}
		}

		// Post the response
		rf.responseCh <- response
	}
}

func (rf *defaultReceiptFetcher) fetchReceipts(block int64, txs []*etypes.Transaction) {
	rf.requestCh <- &txReceiptRequest{blockNumber: block, txs: txs}
}

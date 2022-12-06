package solana

import (
	"math/rand"

	"github.com/ybbus/jsonrpc/v3"
)

// shuffleClients shuffles and returns random permutation of a list of clients.
func shuffleClients(c []jsonrpc.RPCClient) []jsonrpc.RPCClient {
	clients := c
	for i := 0; i < len(clients)*2; i++ {
		x := rand.Intn(len(clients))
		y := rand.Intn(len(clients))
		temp := clients[x]
		clients[x] = clients[y]
		clients[y] = temp
	}

	return clients
}

// executeWithClients tries to execute a function with a list of RPC clients. If any of the execution
// finishes (either with success or failure), the loop through clients list will stop.
// The passed in params f will inform executeWithClients when to stop execution in its return value.
func executeWithClients[T any](originalClients []jsonrpc.RPCClient, f func(client jsonrpc.RPCClient) (T, bool, error)) (T, error) {
	clients := shuffleClients(originalClients)
	var err error
	var stop bool
	var result T
	for _, client := range clients {
		if result, stop, err = f(client); err == nil || stop {
			return result, err
		}
	}

	return result, err
}

package utils

import "github.com/ethereum/go-ethereum/ethclient"

// executeWithClients tries to execute a function with a list of RPC clients. If any of the execution
// finishes (either with success or failure), the loop through clients list will stop.
// The passed in params f will inform executeWithClients when to stop execution in its return value.
func ExecuteWithClients[T any](clients []*ethclient.Client, f func(client *ethclient.Client) (T, bool, error)) (T, error) {
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

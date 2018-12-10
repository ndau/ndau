package tool

import "github.com/tendermint/tendermint/rpc/client"

// Client sets up a client connection to a Tendermint node from its address
//
// The returned type *client.HTTP implements the client.ABCIClient interface
func Client(node string) *client.HTTP {
	return client.NewHTTP(node, "/websocket")
}

package ws

import (
	"errors"

	"github.com/tendermint/tendermint/rpc/client"
)

// Node sets up a client connection to a Tendermint node
func Node(nodeAddress string) (*client.HTTP, error) {
	if nodeAddress == "" {
		return nil, errors.New("node address cannot be empty")
	}
	return client.NewHTTP(nodeAddress, "/websocket"), nil
}

package tool

import (
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Info gets the node's current information and pretty-prints it
func Info(node client.StatusClient) (*rpctypes.ResultStatus, error) {
	status, err := node.Status()
	if err != nil {
		return nil, err
	}
	return status, nil
}

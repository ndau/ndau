package tool

import (
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Version gets the version the connected node is running
func Version(node client.ABCIClient) (
	string, *rpctypes.ResultABCIQuery, error,
) {
	// perform the query
	res, err := node.ABCIQuery(query.VersionEndpoint, []byte{})
	return string(res.Response.Value), res, err
}

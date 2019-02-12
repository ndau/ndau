package tool

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func GetDelegates(node client.ABCIClient) (map[address.Address][]address.Address, *rpctypes.ResultABCIQuery, error) {
	// perform the query
	res, err := node.ABCIQuery(query.DelegatesEndpoint, nil)
	if err != nil {
		return nil, res, err
	}
	if code.ReturnCode(res.Response.Code) != code.OK {
		if res.Response.Log != "" {
			return nil, res, errors.New(code.ReturnCode(res.Response.Code).String() + ": " + res.Response.Log)
		}
		return nil, res, errors.New(code.ReturnCode(res.Response.Code).String())
	}

	dr := query.DelegatesResponse{}
	// parse the response
	_, err = dr.UnmarshalMsg(res.Response.GetValue())
	if err != nil || dr == nil {
		return nil, res, errors.Wrap(err, "unmarshalling delegates response")
	}

	// transform response into friendly form
	delegates := make(map[address.Address][]address.Address)
	for _, node := range dr {
		for _, delegated := range node.Delegated {
			delegates[node.Node] = append(delegates[node.Node], delegated)
		}
	}
	return delegates, res, err
}

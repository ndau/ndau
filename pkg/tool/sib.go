package tool

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetSIB returns the current SIB in effect
func GetSIB(node client.ABCIClient) (
	sib eai.Rate, resp *rpctypes.ResultABCIQuery, err error,
) {
	// perform the query
	resp, err = node.ABCIQuery(query.SIBEndpoint, nil)
	if err != nil {
		return
	}

	// parse the response
	_, err = sib.UnmarshalMsg(resp.Response.Value)
	if err != nil {
		return
	}

	// promote returned errors
	if code.ReturnCode(resp.Response.Code) != code.OK {
		err = errors.New(resp.Response.Log)
		return
	}

	return
}

package tool

import (
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetSummary gets the summary information from the state
func GetSummary(node client.ABCIClient) (*query.Summary, *rpctypes.ResultABCIQuery, error) {
	summ := new(query.Summary)
	// perform the query
	res, err := node.ABCIQuery(query.SummaryEndpoint, nil)
	if err != nil {
		return nil, res, err
	}

	// parse the response
	_, err = summ.UnmarshalMsg(res.Response.GetValue())
	return summ, res, err
}

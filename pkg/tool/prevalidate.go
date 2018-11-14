package tool

import (
	"fmt"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/query"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Prevalidate prevalidates the provided transactable
func Prevalidate(node client.ABCIClient, tx metatx.Transactable) (
	math.Ndau, *rpctypes.ResultABCIQuery, error,
) {
	txb, err := metatx.Marshal(tx, ndau.TxIDs)
	if err != nil {
		return 0, nil, err
	}

	// perform the query
	resp, err := node.ABCIQuery(query.PrevalidateEndpoint, txb)
	if err != nil {
		return 0, resp, err
	}

	// parse the response
	var fee math.Ndau
	_, err = fmt.Sscanf(resp.Response.Info, query.PrevalidateInfoFmt, &fee)
	if err != nil {
		return fee, resp, err
	}

	// promote returned errors
	if code.ReturnCode(resp.Response.Code) != code.OK {
		return fee, resp, errors.New(resp.Response.Log)
	}
	return fee, resp, err
}

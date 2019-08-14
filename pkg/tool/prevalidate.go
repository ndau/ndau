package tool

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/query"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Prevalidate prevalidates the provided transactable
func Prevalidate(node client.ABCIClient, tx metatx.Transactable, logger logrus.FieldLogger) (
	fee math.Ndau, sib math.Ndau, resp *rpctypes.ResultABCIQuery, err error,
) {
	txb, err := metatx.Marshal(tx, ndau.TxIDs)
	if err != nil {
		logger.WithError(err).Info("Prevalidate failed to marshal")
		return
	}

	// perform the query
	resp, err = node.ABCIQuery(query.PrevalidateEndpoint, txb)
	if err != nil {
		logger.WithError(err).Info("Prevalidate query failed")
		return
	}

	// parse the response
	_, err = fmt.Sscanf(resp.Response.Info, query.PrevalidateInfoFmt, &fee, &sib)
	if err != nil {
		l := logger.WithError(err)
		if resp != nil {
			l = l.WithField("response.Info", resp.Response.Info).WithField("response.Log", resp.Response.Log)
		}
		l.Info("Prevalidate sscanf failed")
		// DO NOT RETURN HERE
	}

	// promote returned errors
	if code.ReturnCode(resp.Response.Code) != code.OK {
		err = errors.New(resp.Response.Log)
		return
	}

	return
}

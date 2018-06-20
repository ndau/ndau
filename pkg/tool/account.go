package tool

import (
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
)

// GetAccount gets the account data associated with a given address
func GetAccount(node client.ABCIClient, addrS string) (
	*backing.AccountData, *rpctypes.ResultABCIQuery, error,
) {
	// validate the input address
	addr, err := address.Validate(addrS)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Address validation failure")
	}
	addrB := []byte(addr.String())

	// perform the query
	res, err := node.ABCIQuery(ndau.AccountEndpoint, addrB)
	if err != nil {
		return nil, res, err
	}

	// parse the response
	ad := new(backing.AccountData)
	_, err = ad.UnmarshalMsg(res.Response.GetValue())
	return ad, res, err
}

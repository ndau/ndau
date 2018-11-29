package tool

import (
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// GetAccount gets the account data associated with a given address
func GetAccount(node client.ABCIClient, addr address.Address) (
	*backing.AccountData, *rpctypes.ResultABCIQuery, error,
) {
	addrB := []byte(addr.String())

	// perform the query
	res, err := node.ABCIQuery(query.AccountEndpoint, addrB)
	if err != nil {
		return nil, res, err
	}

	// parse the response
	ad := new(backing.AccountData)
	_, err = ad.UnmarshalMsg(res.Response.GetValue())
	return ad, res, err
}

// GetSequence gets the current sequence number of a particular account
func GetSequence(node client.ABCIClient, addr address.Address) (uint64, error) {
	acct, _, err := GetAccount(node, addr)
	if err != nil {
		return 0, err
	}
	return acct.Sequence, nil
}

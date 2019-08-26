package tool

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
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

// GetAccountHistory gets account data history associated with a given address.
// Pass params as a json-encoded search.AccountHistoryParams object.
func GetAccountHistory(node client.ABCIClient, params search.AccountHistoryParams) (
	*search.AccountHistoryResponse, *rpctypes.ResultABCIQuery, error,
) {
	ps, err := json.Marshal(params)
	if err != nil {
		return nil, nil, errors.Wrap(err, "marshaling account history query params")
	}
	// perform the query
	res, err := node.ABCIQuery(query.AccountHistoryEndpoint, ps)
	if err != nil {
		return nil, res, err
	}

	// parse the response
	ahr := new(search.AccountHistoryResponse)
	err = ahr.Unmarshal(string(res.Response.GetValue()))
	return ahr, res, err
}

// GetAccountList gets a list of account names, paged according to the params
// Pass in after = "" (which is less than all nonempty strings) and limit = 0
// to get all results. (Note that the ndauapi will enforce a limit of 100 items.)
func GetAccountList(node client.ABCIClient, after string, limit int) (
	*query.AccountListQueryResponse, *rpctypes.ResultABCIQuery, error,
) {
	// Prepare search params.
	params := search.AccountListParams{
		Address: "",
		After:   after,
		Limit:   limit,
	}
	paramsBuf, err := json.Marshal(params)
	if err != nil {
		return nil, nil, err
	}

	// perform the query
	res, err := node.ABCIQuery(query.AccountListEndpoint, paramsBuf)
	if err != nil {
		return nil, res, err
	}

	// parse the response
	var result query.AccountListQueryResponse
	_, err = result.UnmarshalMsg(res.Response.GetValue())
	return &result, res, err
}

// GetAccountListBatch abstracts over the process of repeatedly calling
// GetAccountList in order to get a complete list of all known addresses.
//
// This function makes a best-effort attempt to return a complete and current
// list of accounts known to the node, but true consistency is impossible using
// a sequential paged API; as we cannot lock the node, there may be updates
// during paging which cause addresses to appear in pages we have already
// visited. This is unavoidable.
func GetAccountListBatch(node client.ABCIClient) ([]address.Address, error) {
	var (
		accts = make([]string, 0)
		after = ""
		limit = 100

		qaccts *query.AccountListQueryResponse
		err    error
	)

	getPage := func() {
		qaccts, _, err = GetAccountList(
			node,
			after,
			limit,
		)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf(
				"getPage(%s, %d)", after, limit,
			))
			return
		}
		accts = append(accts, qaccts.Accounts...)
		after = qaccts.NextAfter
	}

	// prime the pump
	getPage()
	if err != nil {
		return nil, err
	}
	for after != "" {
		getPage()
		if err != nil {
			return nil, err
		}
	}

	// eliminate duplicate accts and convert to address type
	sort.Strings(accts)
	addrs := make([]address.Address, 0, len(accts))
	for _, acct := range accts {
		addr, err := address.Validate(acct)
		if err != nil {
			return nil, errors.Wrap(err, "GetAccountListBatch validating acct addr")
		}
		if len(addrs) == 0 || addr != addrs[len(addrs)-1] {
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

// GetCurrencySeats gets a list of ndau currency seats
//
// Currency seats are defined as those accounts containing more than 1000 ndau.
func GetCurrencySeats(node client.ABCIClient) ([]address.Address, error) {
	addrs, err := GetAccountListBatch(node)
	if err != nil {
		return nil, errors.Wrap(err, "GetCurrencySeats")
	}
	seats := make(map[address.Address]*backing.AccountData)
	for _, addr := range addrs {
		data, _, err := GetAccount(node, addr)
		if err != nil {
			return nil, errors.Wrap(err, "GetCurrencySeats")
		}
		if data.CurrencySeatDate != nil {
			seats[addr] = data
		}
	}
	seatAddrs := make([]address.Address, 0, len(seats))
	for seat := range seats {
		seatAddrs = append(seatAddrs, seat)
	}
	// sort seats by currency seat date, oldest first
	// this way, getting the oldest `N` is as simple as slicing: `[:N]`
	sort.Slice(seatAddrs, func(i, j int) bool {
		return *seats[seatAddrs[i]].CurrencySeatDate < *seats[seatAddrs[j]].CurrencySeatDate
	})
	return seatAddrs, err
}

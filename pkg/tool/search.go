package tool

import (
	"encoding/json"

	metasrch "github.com/oneiro-ndev/metanode/pkg/meta/search"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

// GetSearchResults returns search results for a given search query.
func GetSearchResults(node client.ABCIClient, params search.QueryParams) (
	string, error,
) {
	// encode the query
	ahpj, err := json.Marshal(params)
	if err != nil {
		return "", errors.Wrap(err, "marshaling params")
	}

	// perform the query
	res, err := node.ABCIQuery(query.SearchEndpoint, ahpj)
	if err != nil {
		return "", errors.Wrap(err, "performing query")
	}

	// parse the response
	searchValue := string(res.Response.GetValue())
	return searchValue, nil
}

// SearchDateRange returns the first and last block heights for the given ISO-3339 date range.
func SearchDateRange(node client.ABCIClient, first, last string) (
	uint64, uint64, error,
) {
	request := metasrch.DateRangeRequest{FirstTimestamp: first, LastTimestamp: last}

	// perform the query
	res, err := node.ABCIQuery(query.DateRangeEndpoint, []byte(request.Marshal()))
	if err != nil {
		return 0, 0, err
	}

	// parse the response
	searchValue := string(res.Response.GetValue())
	var result metasrch.DateRangeResult
	result.Unmarshal(searchValue)
	return result.FirstHeight, result.LastHeight, nil
}

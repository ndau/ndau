package tool

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
)

// GetSearchResults returns search results for a given search query.
// Pass params as a json-encoded search.QueryParams object.
func GetSearchResults(node client.ABCIClient, params string) (
	string, error,
) {
	// perform the query
	res, err := node.ABCIQuery(query.SearchEndpoint, []byte(params))
	if err != nil {
		return "", err
	}

	// parse the response
	searchValue := string(res.Response.GetValue())
	return searchValue, nil
}

// SearchDateRange returns the first and last block heights for the given ISO-3339 date range.
func SearchDateRange(node client.ABCIClient, first, last string) (
	uint64, uint64, error,
) {
	request := search.DateRangeRequest{FirstTimestamp: first, LastTimestamp: last}

	// perform the query
	res, err := node.ABCIQuery(query.DateRangeEndpoint, []byte(request.Marshal()))
	if err != nil {
		return 0, 0, err
	}

	// parse the response
	searchValue := string(res.Response.GetValue())
	var result search.DateRangeResult
	result.Unmarshal(searchValue)
	return result.FirstHeight, result.LastHeight, nil
}

package tool

import (
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
)

// GetSearchResults returns search results for a given search query.
// Pass params in the format "a=b&c=d&...&y=z"
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
	return searchValue, err
}

package tool

import (
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
)

// GetSearchResults returns search results for a given search query.
// Pass params in the format "a=b&c=d&...&y=z"
func GetSearchResults(node client.ABCIClient, params string) (
	int64, error,
) {
	// perform the query
	res, err := node.ABCIQuery(query.SearchEndpoint, []byte(params))
	if err != nil {
		return 0, err
	}

	// parse the response
	value := string(res.Response.GetValue())
	height, err := strconv.ParseInt(value, 10, 64)
	return height, err
}

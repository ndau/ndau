package tool

import (
	srch "github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

func priceHistory(
	node client.ABCIClient,
	endpoint string,
	params srch.PriceQueryParams,
) (srch.PriceQueryResults, error) {
	var out srch.PriceQueryResults
	pqpb, err := params.MarshalMsg(nil)
	if err != nil {
		return out, errors.Wrap(err, "marshaling params")
	}
	resp, err := node.ABCIQuery(endpoint, pqpb)
	if err != nil {
		return out, errors.Wrap(err, "performing query")
	}
	_, err = out.UnmarshalMsg(resp.Response.Value)
	err = errors.Wrap(err, "unmarshaling response")
	return out, err
}

// TargetPriceHistory returns historical data for the ndau target price
func TargetPriceHistory(
	node client.ABCIClient,
	params srch.PriceQueryParams,
) (srch.PriceQueryResults, error) {
	return priceHistory(node, query.PriceTargetEndpoint, params)
}

// MarketPriceHistory returns historical data for the ndau market price
func MarketPriceHistory(
	node client.ABCIClient,
	params srch.PriceQueryParams,
) (srch.PriceQueryResults, error) {
	return priceHistory(node, query.PriceMarketEndpoint, params)
}

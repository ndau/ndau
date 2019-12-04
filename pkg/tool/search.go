package tool

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"encoding/json"

	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/query"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
)

// GetSearchResults returns search results for a given search query.
func GetSearchResults(node client.ABCIClient, params search.QueryParams) (
	[]byte, error,
) {
	// encode the query
	ahpj, err := json.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling params")
	}

	// perform the query
	res, err := node.ABCIQuery(query.SearchEndpoint, ahpj)
	if err != nil {
		return nil, errors.Wrap(err, "performing query")
	}

	return res.Response.GetValue(), nil
}

// SearchDateRange returns the first and last block heights for the given date range.
func SearchDateRange(node client.ABCIClient, first, last math.Timestamp) (
	uint64, uint64, error,
) {
	request := query.DateRangeRequest{FirstTimestamp: first, LastTimestamp: last}

	// perform the query
	data, err := request.MarshalMsg(nil)
	if err != nil {
		return 0, 0, errors.Wrap(err, "marshaling request")
	}
	res, err := node.ABCIQuery(query.DateRangeEndpoint, data)
	if err != nil {
		return 0, 0, errors.Wrap(err, "querying date range endpoint")
	}

	// parse the response
	var result query.DateRangeResult
	_, err = result.UnmarshalMsg(res.Response.Value)
	return result.FirstHeight, result.LastHeight, nil
}

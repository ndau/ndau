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
	"github.com/ndau/metanode/pkg/meta/app/code"
	"github.com/ndau/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetSummary gets the summary information from the state
func GetSummary(node client.ABCIClient) (*query.Summary, *rpctypes.ResultABCIQuery, error) {
	summ := new(query.Summary)
	// perform the query
	res, err := node.ABCIQuery(query.SummaryEndpoint, nil)
	if err != nil {
		return nil, res, err
	}
	if code.ReturnCode(res.Response.Code) != code.OK {
		return nil, res, errors.New(code.ReturnCode(res.Response.Code).String() + ": " + res.Response.Log)
	}

	// parse the response
	_, err = summ.UnmarshalMsg(res.Response.GetValue())
	return summ, res, err
}

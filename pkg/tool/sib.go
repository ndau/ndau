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
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetSIB returns the current SIB in effect
func GetSIB(node client.ABCIClient) (
	sib query.SIBResponse, resp *rpctypes.ResultABCIQuery, err error,
) {
	// perform the query
	resp, err = node.ABCIQuery(query.SIBEndpoint, nil)
	if err != nil {
		return
	}

	// parse the response
	_, err = sib.UnmarshalMsg(resp.Response.Value)
	if err != nil {
		return
	}

	// promote returned errors
	if code.ReturnCode(resp.Response.Code) != code.OK {
		err = errors.New(resp.Response.Log)
		return
	}

	return
}

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
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetDelegates gets the set of nodes with delegates, and the list of accounts delegated to each
func GetDelegates(node client.ABCIClient) (map[address.Address][]address.Address, *rpctypes.ResultABCIQuery, error) {
	// perform the query
	res, err := node.ABCIQuery(query.DelegatesEndpoint, nil)
	if err != nil {
		return nil, res, err
	}
	if code.ReturnCode(res.Response.Code) != code.OK {
		if res.Response.Log != "" {
			return nil, res, errors.New(code.ReturnCode(res.Response.Code).String() + ": " + res.Response.Log)
		}
		return nil, res, errors.New(code.ReturnCode(res.Response.Code).String())
	}

	dr := query.DelegatesResponse{}
	// parse the response
	_, err = dr.UnmarshalMsg(res.Response.GetValue())
	if err != nil || dr == nil {
		return nil, res, errors.Wrap(err, "unmarshalling delegates response")
	}

	// transform response into friendly form
	delegates := make(map[address.Address][]address.Address)
	for _, node := range dr {
		for _, delegated := range node.Delegated {
			delegates[node.Node] = append(delegates[node.Node], delegated)
		}
	}
	return delegates, res, err
}

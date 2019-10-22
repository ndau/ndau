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
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Version gets the version the connected node is running
func Version(node client.ABCIClient) (
	string, *rpctypes.ResultABCIQuery, error,
) {
	// perform the query
	res, err := node.ABCIQuery(query.VersionEndpoint, []byte{})
	if err != nil {
		return "", res, err
	}
	return string(res.Response.GetValue()), res, err
}

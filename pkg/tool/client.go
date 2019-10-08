package tool

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import "github.com/tendermint/tendermint/rpc/client"

// Client sets up a client connection to a Tendermint node from its address
//
// The returned type *client.HTTP implements the client.ABCIClient interface
func Client(node string) *client.HTTP {
	return client.NewHTTP(node, "/websocket")
}

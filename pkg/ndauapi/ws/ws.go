package ws

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"errors"

	"github.com/oneiro-ndev/tendermint.0.32.3/rpc/client"
)

// Node sets up a client connection to a Tendermint node
func Node(nodeAddress string) (*client.HTTP, error) {
	if nodeAddress == "" {
		return nil, errors.New("node address cannot be empty")
	}
	return client.NewHTTP(nodeAddress, "/websocket"), nil
}

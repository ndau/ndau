package backing

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import math "github.com/ndau/ndaumath/pkg/types"

// generate msgp interface implementations for AccountData and supporting structs
// we can't generate the streaming interfaces, unfortunately, because the
// signature.* types don't implement those
//go:generate msgp -io=0

// generate noms marshaler implementations for appropriate types
//go:generate go run $GOPATH/src/github.com/ndau/generator/cmd/nomsify $GOPATH/src/github.com/ndau/ndau/pkg/ndau/backing
//go:generate find $GOPATH/src/github.com/ndau/ndau/pkg/ndau/backing -name "*noms_gen*.go" -maxdepth 1 -exec goimports -w {} ;
//nomsify Hold

// Hold tracks a portion of this account's ndau which cannot currently be spent
type Hold struct {
	Qty    math.Ndau       `json:"qty" chain:"81,Hold_Quantity"`
	Expiry *math.Timestamp `json:"expiry" chain:"82,Hold_Expiry"`
	Txhash string          `json:"tx_hash" chain:"83,Hold_TxHash"`
	Stake  *StakeData      `json:"stake" chain:"84,Hold_Stake"`
}

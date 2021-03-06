package svc

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"net/http"
	"testing"

	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/stretchr/testify/assert"
)

func TestRouting(t *testing.T) {
	type rt struct {
		verb string
		in   string
		out  string
	}

	routes := []rt{
		rt{"GET", "/account/account/123456", "/account/account/:address"},
		rt{"POST", "/account/accounts", "/account/accounts"},
		rt{"GET", "/account/history/123456", "/account/history/:address"},
		rt{"GET", "/account/list", "/account/list"},
		rt{"GET", "/account/currencyseats", "/account/currencyseats"},
		rt{"GET", "/block/before/123", "/block/before/:height"},
		rt{"GET", "/block/hash/abc123", "/block/hash/:blockhash"},
		rt{"GET", "/block/height/10234", "/block/height/:height"},
		rt{"GET", "/block/range/123/143", "/block/range/:first/:last"},
		rt{"GET", "/block/daterange/x/y", "/block/daterange/:first/:last"},
		rt{"GET", "/block/transactions/555", "/block/transactions/:height"},
		rt{"GET", "/node/status", "/node/status"},
		rt{"GET", "/node/health", "/node/health"},
		rt{"GET", "/node/net", "/node/net"},
		rt{"GET", "/node/genesis", "/node/genesis"},
		rt{"GET", "/node/abci", "/node/abci"},
		rt{"GET", "/node/consensus", "/node/consensus"},
		rt{"GET", "/node/nodes", "/node/nodes"},
		rt{"GET", "/node/registerednodes", "/node/registerednodes"},
		rt{"GET", "/node/ad349f", "/node/:id"},
		rt{"POST", "/price/target/history", "/price/target/history"},
		rt{"POST", "/price/market/history", "/price/market/history"},
		rt{"GET", "/price/current", "/price/current"},
		rt{"GET", "/system/all", "/system/all"},
		rt{"GET", "/system/get/foo,bar", "/system/get/:sysvars"},
		rt{"POST", "/system/set/foo", "/system/set/:sysvar"},
		rt{"GET", "/system/history/foo", "/system/history/:sysvar"},
		rt{"POST", "/system/eai/rate", "/system/eai/rate"},
		rt{"GET", "/transaction/detail/5469abfed", "/transaction/detail/:txhash"},
		rt{"GET", "/transaction/before/5469abfed", "/transaction/before/:txhash"},
		rt{"POST", "/tx/prevalidate/lock", "/tx/prevalidate/:txtype"},
		rt{"POST", "/tx/submit/transfer", "/tx/submit/:txtype"},
		rt{"GET", "/version", "/version"},
	}

	cf := cfg.Cfg{}
	mux := New(cf).Mux()

	for _, r := range routes {
		t.Run(r.out, func(t *testing.T) {
			req, _ := http.NewRequest(r.verb, r.in, nil)
			route := mux.GetRequestRoute(req)
			assert.Equal(t, r.out, route)
		})
	}
}

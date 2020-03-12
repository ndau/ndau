package routes

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"
	"net/http"

	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
)

// GetGenesis returns the genesis doc from tendermint.
func GetGenesis(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		genesis, err := cf.Node.Genesis()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Could not fetch genesis: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(genesis))
	}
}

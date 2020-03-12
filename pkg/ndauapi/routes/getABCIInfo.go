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

// GetABCIInfo proxies tendermint's ABCIInfo
func GetABCIInfo(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		abciInfo, err := cf.Node.ABCIInfo()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not fetch ABCI info: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(abciInfo))
	}
}

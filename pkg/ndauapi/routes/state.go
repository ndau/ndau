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
	"github.com/ndau/ndau/pkg/tool"
)

// HandleStateDelegates returns a HandlerFunc that returns all the delegates
// in the system
func HandleStateDelegates(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delegates, _, err := tool.GetDelegates(cf.Node)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching delegate list: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(delegates))
	}
}

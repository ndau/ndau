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

// HealthStatus gives us the ability to add more status information later without messing up clients
type HealthStatus struct {
	Status string
}

// HealthResponse is the response from the /health endpoint.
type HealthResponse struct {
	Ndau HealthStatus
}

// GetHealth returns health indicators from Tendermint.
func GetHealth(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// the Health function returns a null object, so if it doesn't error we're good
		_, err := cf.Node.Health()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Could not fetch ndau node health: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(HealthResponse{HealthStatus{"Ok"}}))
	}
}

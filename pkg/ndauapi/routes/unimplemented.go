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
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// PriceHistoryRecord struct
type PriceHistoryRecord struct {
	Timestamp math.Timestamp
	PriceData PriceInfo
}

// HandleNumUnconfirmedTxs func
func HandleNumUnconfirmedTxs(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandlePriceHeight func
func HandlePriceHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandlePriceHistory func
func HandlePriceHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

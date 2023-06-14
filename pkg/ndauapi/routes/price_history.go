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
	"encoding/json"
	"io/ioutil"
	"net/http"

	srch "github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/oneiro-ndev/tendermint.0.32.3/rpc/client"
)

// PriceHistoryResults encapsulates a set of price history data in a json-friendly way
type PriceHistoryResults struct {
	Items []srch.PriceQueryResult `json:"items"`
	Next  string                  `json:"next"`
}

// HandlePriceTargetHistory handles price target history
func HandlePriceTargetHistory(cf cfg.Cfg) http.HandlerFunc {
	return priceHistory(cf, tool.TargetPriceHistory)
}

// HandlePriceMarketHistory handles price market history
func HandlePriceMarketHistory(cf cfg.Cfg) http.HandlerFunc {
	return priceHistory(cf, tool.MarketPriceHistory)
}

func priceHistory(
	cf cfg.Cfg,
	tf func(
		node client.ABCIClient,
		params srch.PriceQueryParams,
	) (srch.PriceQueryResults, error),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bdata, err := ioutil.ReadAll(r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("reading params", err, http.StatusBadRequest))
			return
		}

		var params srch.PriceQueryParams
		err = json.Unmarshal(bdata, &params)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unmarshaling params", err, http.StatusBadRequest))
			return
		}

		if params.Limit == 0 {
			params.Limit = 100
		}
		if params.Limit > 1000 {
			params.Limit = 1000
		}

		pqr, err := tf(cf.Node, params)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("searching history", err, http.StatusInternalServerError))
			return
		}

		out := PriceHistoryResults{
			Items: pqr.Items,
		}
		if pqr.More && len(pqr.Items) > 0 {
			params.After = srch.RangeEndpoint{Timestamp: pqr.Items[len(pqr.Items)-1].Timestamp}
			data, err := json.Marshal(params)
			if err == nil {
				// otherwise, just forget it; this is a convenience, not essential
				out.Next = string(data)
			}
		}

		reqres.RespondJSON(w, reqres.OKResponse(out))
	}
}

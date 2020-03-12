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
	"time"

	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/types"
)

// EAIRateRequest is the type of a single instance of the rate request (the API takes
// an array).
type EAIRateRequest struct {
	Address string          `json:"address"`
	WAA     types.Duration  `json:"weightedAverageAge"`
	Lock    backing.Lock    `json:"lock"`
	At      types.Timestamp `json:"at"`
}

// EAIRateResponse is a single instance of a rate response (it returns an array of them)
type EAIRateResponse struct {
	Address string `json:"address"`
	EAIRate uint64 `json:"eairate"`
}

// GetEAIRate returns the EAI rates for a collection of rate requests, each of which has
// an address (merely a string that is not examined, simply copied to the output), a
// weighted average age, and optional lock information (if the account is locked)
func GetEAIRate(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requests []EAIRateRequest
		if r.Body == nil {
			reqres.RespondJSON(w, reqres.NewAPIError("request body required", http.StatusBadRequest))
			return
		}
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("failed to read request body", err, http.StatusBadRequest))
			return
		}
		err = json.Unmarshal(buf, &requests)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unable to decode", err, http.StatusBadRequest))
			return
		}

		// FIXME: need to actually query the chaos chain
		// These are just the default values
		unlockedTable := eai.DefaultUnlockedEAI

		// Below is code that came from the ndau app but can't run because we don't have
		// app or config objects.

		// sc, err := cache.NewSystemCache(config)
		// if err != nil {
		// 	reqres.RespondJSON(w, reqres.NewFromErr("couldn't create cache", err, http.StatusInternalServerError))
		// 	return
		// }

		// err = app.System(sv.UnlockedRateTableName, unlockedTable)
		// if err != nil {
		// 	return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
		// }

		now, err := types.TimestampFrom(time.Now())
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("cannot get current time", err, http.StatusInternalServerError))
			return
		}

		response := make([]EAIRateResponse, len(requests))
		for i := range requests {
			response[i].Address = requests[i].Address
			if requests[i].At == 0 {
				requests[i].At = now
			}
			response[i].EAIRate = uint64(
				eai.CalculateEAIRate(requests[i].WAA, &requests[i].Lock, unlockedTable, requests[i].At),
			)
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

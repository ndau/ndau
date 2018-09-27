package routes

import (
	"encoding/json"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// EAIRateRequest is the type of a single instance of the rate request (the API takes
// an array).
type EAIRateRequest struct {
	Address string         `json:"address"`
	WAA     types.Duration `json:"weightedAverageAge"`
	Lock    backing.Lock   `json:"lock"`
}

// EAIRateResponse is a single instance of a rate response (it returns an array of them)
type EAIRateResponse struct {
	Address string `json:"address"`
	EAIRate int64  `json:"eairate"`
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
		err := json.NewDecoder(r.Body).Decode(&requests)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unable to decode", err, http.StatusBadRequest))
			return
		}

		// TODO: need to actually query the chaos chain
		// These are just the default values
		unlockedTable := eai.DefaultUnlockedEAI
		lockedTable := eai.DefaultLockBonusEAI

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
		// lockedTable := new(eai.RateTable)
		// err = app.System(sv.LockedRateTableName, lockedTable)
		// if err != nil {
		// 	return errors.Wrap(err, fmt.Sprintf("Error fetching %s system variable in CreditEAI.Apply", sv.UnlockedRateTableName))
		// }

		response := make([]EAIRateResponse, len(requests))
		for i := range requests {
			response[i].Address = requests[i].Address
			response[i].EAIRate = eai.CalculateEAIRate(requests[i].WAA, &requests[i].Lock, unlockedTable, lockedTable)
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

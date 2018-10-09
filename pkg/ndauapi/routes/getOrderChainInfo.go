package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
)

// OrderChainInfo is a single instance of a rate response (it returns an array of them)
type OrderChainInfo struct {
	MarketPrice   float64 `json:"marketPrice"`
	TargetPrice   float64 `json:"targetPrice"`
	FloorPrice    float64 `json:"floorPrice"`
	EndowmentSold int64   `json:"endowmentSold"`
	TotalNdau     int64   `json:"totalNdau"`
	PriceUnits    string  `json:"USD"`
}

func getOrderChainInfo() (OrderChainInfo, error) {
	info := OrderChainInfo{
		MarketPrice:   16.85,
		TargetPrice:   17.00,
		FloorPrice:    2.57,
		EndowmentSold: 2919000 * 100000000,
		TotalNdau:     3141593 * 100000000,
		PriceUnits:    "USD",
	}
	return info, nil
}

// GetOrderChainData returns a block of information for the
func GetOrderChainData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		response, err := getOrderChainInfo()
		if err != nil {
			reqres.NewFromErr("couldn't query the order chain", err, http.StatusInternalServerError)
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

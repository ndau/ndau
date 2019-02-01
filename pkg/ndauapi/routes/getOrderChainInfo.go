package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// OrderChainInfo is a single instance of a rate response (it returns an array of them)
type OrderChainInfo struct {
	MarketPrice   float64    `json:"marketPrice"`
	TargetPrice   float64    `json:"targetPrice"`
	FloorPrice    float64    `json:"floorPrice"`
	EndowmentSold types.Ndau `json:"endowmentSold"`
	TotalNdau     types.Ndau `json:"totalNdau"`
	SIB           float64    `json:"sib"`
	PriceUnits    string     `json:"priceUnit"`
}

func getTotals(cf cfg.Cfg) (types.Ndau, types.Ndau, error) {
	node, err := ws.Node(cf.NodeAddress)
	if err != nil {
		return 0, 0, err
	}
	summ, _, err := tool.GetSummary(node)
	if err != nil {
		return 0, 0, err
	}
	return summ.TotalNdau, summ.EndowmentSold, nil
}

// The idea behind floor price is that even if you sell off all the ndau in the world
// at the floor price, you can't drain away more than half of the endowment's value
// The floor price is the total value of the endowment divided by the total ndau in
// circulation, divided by two.
// The total value of the endowment is fetched from a system variable -- however,
// if that system variable is not defined, we use the total purchase price of all
// ndau instead.
func getFloorPrice(totalNdau types.Ndau) {
}

func getSIB(targetPrice, marketPrice, floorPrice float64) float {
	target95 := targetPrice * 0.95
	// for safety reasons, we check to make sure floor price is reasonable; it should never
	// get this high
	if marketPrice >= target95 || floorPrice >= target95 {
		return 1.0
	}
	return 0.5 * (target95 - marketPrice) / (target95 - floorPrice)
}

func getOrderChainInfo(cf cfg.Cfg) (OrderChainInfo, error) {
	totalndau, totalRFE, err := getTotals(cf)
	targetPrice := pricecurve.PriceAtUnit(totalRFE)
	// floorPrice = getFloorPrice(totalRFE, endowmentValue)
	info := OrderChainInfo{
		MarketPrice:   targetPrice,
		TargetPrice:   targetPrice,
		FloorPrice:    2.57,
		EndowmentSold: totalRFE,
		TotalNdau:     totalndau,
		SIB:           0,
		PriceUnits:    "USD",
	}
	return info, err
}

// GetOrderChainData returns a block of information from the order chain
// (Although for now it's mocked up to return fake data)
func GetOrderChainData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getOrderChainInfo(cf)
		if err != nil {
			reqres.NewFromErr("order chain query error", err, http.StatusInternalServerError)
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

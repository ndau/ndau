package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// OrderChainInfo is a single instance of a rate response (it returns an array of them)
type OrderChainInfo struct {
	MarketPrice float64    `json:"marketPrice"`
	TargetPrice float64    `json:"targetPrice"`
	FloorPrice  float64    `json:"floorPrice"`
	TotalIssued types.Ndau `json:"totalIssued"`
	TotalNdau   types.Ndau `json:"totalNdau"`
	SIB         float64    `json:"sib"`
	PriceUnits  string     `json:"priceUnit"`
}

// getTotals builds the start of an OrderChainInfo object, filling in the basics
func getTotals(cf cfg.Cfg) (OrderChainInfo, error) {
	var oci OrderChainInfo
	node, err := ws.Node(cf.NodeAddress)
	if err != nil {
		return oci, err
	}
	summ, _, err := tool.GetSummary(node)
	if err != nil {
		return oci, err
	}
	oci.TotalIssued = summ.TotalIssue
	oci.TotalNdau = summ.TotalCirculation
	return oci, nil
}

func getSIB(targetPrice, marketPrice, floorPrice float64) float64 {
	target95 := targetPrice * 0.95
	// for safety reasons, we check to make sure floor price is reasonable; it should never
	// get this high
	if marketPrice >= target95 || floorPrice >= target95 {
		return 1.0
	}
	return 0.5 * (target95 - marketPrice) / (target95 - floorPrice)
}

func getAsFloat64(sysvars map[string][]byte, name string, def float64) float64 {
	_, ok := sysvars[name]
	if !ok {
		return def
	}
	// we have to decode bytes here but don't know how yet
	return def
}

// The total value of the endowment and current market price are fetched from system variables.
// If those system variables are not defined or equal to 0, we use:
// * For endowment value -- the total purchase price of all ndau
// * For market price -- the target price.
func getOrderChainInfo(cf cfg.Cfg) (OrderChainInfo, error) {
	oci, err := getTotals(cf)

	sysvars, err := getSystemVars(cf.NodeAddress)
	if err != nil {
		return oci, err
	}
	defaultEndowmentValue := pricecurve.TotalPriceFor(oci.TotalIssued, 0)
	endowmentvalue := getAsFloat64(sysvars, "EndowmentDollarValue", defaultEndowmentValue)

	targetPrice := pricecurve.PriceAtUnit(oci.TotalIssued)
	marketPrice := getAsFloat64(sysvars, "MarketPrice", targetPrice)

	oci.MarketPrice = marketPrice
	oci.TargetPrice = targetPrice
	// The idea behind floor price is that even if you sell off all the ndau in the world
	// at the floor price, you can't drain away more than half of the endowment's value
	// The floor price is the total value of the endowment divided by the total ndau in
	// circulation, divided by two.
	numNdau := float64(oci.TotalNdau) / constants.QuantaPerUnit
	if numNdau != 0 {
		oci.FloorPrice = (endowmentvalue / numNdau) / 2.0
	}
	oci.SIB = getSIB(targetPrice, marketPrice, oci.FloorPrice)
	oci.PriceUnits = "USD"

	return oci, err
}

// GetOrderChainData returns a block of information from the order chain
// (Although for now it's mocked up to return fake data)
func GetOrderChainData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getOrderChainInfo(cf)
		if err != nil {
			reqres.NewFromErr("order chain query error", err, http.StatusInternalServerError)
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

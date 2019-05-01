package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// PriceInfo is a single instance of a rate response (it returns an array of them)
type PriceInfo struct {
	MarketPrice pricecurve.Nanocent `json:"marketPrice"`
	TargetPrice pricecurve.Nanocent `json:"targetPrice"`
	TotalIssued types.Ndau          `json:"totalIssued"`
	TotalNdau   types.Ndau          `json:"totalNdau"`
	TotalSIB    types.Ndau          `json:"totalSIB"`
	CurrentSIB  eai.Rate            `json:"sib"`
}

// getPriceInfo builds a PriceInfo object
func getPriceInfo(cf cfg.Cfg) (PriceInfo, error) {
	var oci PriceInfo
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
	oci.TotalSIB = summ.TotalBurned

	sib, _, err := tool.GetSIB(node)
	if err != nil {
		return oci, err
	}

	oci.CurrentSIB = sib.SIB
	oci.TargetPrice = sib.TargetPrice
	oci.MarketPrice = sib.MarketPrice
	return oci, err
}

// GetPriceData returns a block of price information
func GetPriceData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getPriceInfo(cf)
		if err != nil {
			reqres.NewFromErr("price query error", err, http.StatusInternalServerError)
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

// OrderChainInfo is the old structure returned by /order/price. We only need it as long as we
// keep the deprecated function below.
type OrderChainInfo struct {
	MarketPrice float64    `json:"marketPrice"`
	TargetPrice float64    `json:"targetPrice"`
	FloorPrice  float64    `json:"floorPrice"`
	TotalIssued types.Ndau `json:"totalIssued"`
	TotalNdau   types.Ndau `json:"totalNdau"`
	SIB         float64    `json:"sib"`
	PriceUnits  string     `json:"priceUnit"`
}

// GetOrderData returns a block of price information, but converts it to a format that is compatible
// with the old /order/current API. This is for use while v1.8 of the wallet is still extant.
func GetOrderData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pricedata, err := getPriceInfo(cf)
		if err != nil {
			reqres.NewFromErr("price query error", err, http.StatusInternalServerError)
			return
		}
		oci := OrderChainInfo{
			// converting from nanocents to floating point dollars means multiplying by 10^11
			MarketPrice: float64(pricedata.MarketPrice) / 100000000000.0,
			TargetPrice: float64(pricedata.TargetPrice) / 100000000000.0,
			FloorPrice:  2.40,
			TotalIssued: pricedata.TotalIssued,
			TotalNdau:   pricedata.TotalNdau,
			SIB:         1,
			PriceUnits:  "USD",
		}

		reqres.RespondJSON(w, reqres.OKResponse(oci))
	}
}

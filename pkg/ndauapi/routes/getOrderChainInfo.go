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

// OrderChainInfo is a single instance of a rate response (it returns an array of them)
type OrderChainInfo struct {
	MarketPrice pricecurve.Nanocent `json:"marketPrice"`
	TargetPrice pricecurve.Nanocent `json:"targetPrice"`
	TotalIssued types.Ndau          `json:"totalIssued"`
	TotalNdau   types.Ndau          `json:"totalNdau"`
	TotalSIB    types.Ndau          `json:"totalSIB"`
	CurrentSIB  eai.Rate            `json:"sib"`
}

// getOCI builds the start of an OrderChainInfo object, filling in the basics
func getOCI(cf cfg.Cfg) (OrderChainInfo, error) {
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
	oci.TotalSIB = summ.TotalSIB

	sib, _, err := tool.GetSIB(node)
	if err != nil {
		return oci, err
	}

	oci.CurrentSIB = sib.SIB
	oci.TargetPrice = sib.TargetPrice
	oci.MarketPrice = sib.MarketPrice
	return oci, err
}

// GetOrderChainData returns a block of information from the order chain
func GetOrderChainData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getOCI(cf)
		if err != nil {
			reqres.NewFromErr("order chain query error", err, http.StatusInternalServerError)
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
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

func getTotalNdau(cf cfg.Cfg) (int64, error) {
	node, err := ws.Node(cf.NodeAddress)
	if err != nil {
		return 0, err
	}
	summ, _, err := tool.GetSummary(node)
	return int64(summ.TotalNdau), nil
}

func getOrderChainInfo(cf cfg.Cfg) (OrderChainInfo, error) {
	totalndau, err := getTotalNdau(cf)
	info := OrderChainInfo{
		MarketPrice:   16.85,
		TargetPrice:   17.00,
		FloorPrice:    2.57,
		EndowmentSold: 2919000 * 100000000,
		TotalNdau:     totalndau,
		PriceUnits:    "USD",
	}
	return info, err
}

// GetOrderChainData returns a block of information for the
func GetOrderChainData(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := getOrderChainInfo(cf)
		if err != nil {
			reqres.NewFromErr("order chain query error", err, http.StatusInternalServerError)
		}
		reqres.RespondJSON(w, reqres.OKResponse(response))
	}
}

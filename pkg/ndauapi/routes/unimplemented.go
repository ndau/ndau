package routes

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

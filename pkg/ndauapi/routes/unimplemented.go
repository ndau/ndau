package routes

import (
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// BalanceHistoryItem struct
type BalanceHistoryItem struct {
	Timestamp string
	Balance   math.Ndau
	TxHash    string
}

// ChaosHistoryResponse struct
type ChaosHistoryResponse struct {
}

// OrderHistoryRecord struct
type OrderHistoryRecord struct {
	Timestamp math.Timestamp
	OrderInfo OrderChainInfo
}

// HandleNumUnconfirmedTxs func
func HandleNumUnconfirmedTxs(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandleOrderCurrent func
func HandleOrderCurrent(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandleOrderHash func
func HandleOrderHash(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandleOrderHeight func
func HandleOrderHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

// HandleOrderHistory func
func HandleOrderHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqres.RespondJSON(w, reqres.NewAPIError("unimplemented endpoint", http.StatusNotImplemented))
	}
}

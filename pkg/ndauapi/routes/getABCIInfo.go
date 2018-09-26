package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// GetABCIInfo proxies tendermint's ABCIInfo
func GetABCIInfo(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not get node client", http.StatusInternalServerError))
			return
		}

		abciInfo, err := node.ABCIInfo()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not fetch ABCI info: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(abciInfo))
	}
}

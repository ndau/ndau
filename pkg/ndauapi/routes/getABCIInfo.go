package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
)

// GetABCIInfo proxies tendermint's ABCIInfo
func GetABCIInfo(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		abciInfo, err := cf.Node.ABCIInfo()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not fetch ABCI info: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(abciInfo))
	}
}

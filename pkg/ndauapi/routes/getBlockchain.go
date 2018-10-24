package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// GetBlockchain returns the blockchain within a range.
func GetBlockchain(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := processBlockchainRequest(r)
		if err != nil {
			// Anything that errors from here is going to be a bad request.
			reqres.RespondJSON(w, reqres.NewAPIError(err.Error(), http.StatusBadRequest))
			return
		}
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		block, err := node.BlockchainInfo(req.start, req.end)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(block))
	}
}

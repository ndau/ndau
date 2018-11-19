package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// HealthStatus gives us the ability to add more status information later without messing up clients
type HealthStatus struct {
	Status string
}

// HealthResponse is the response from the /health endpoint.
type HealthResponse struct {
	Chaos HealthStatus
	Ndau  HealthStatus
}

// GetHealth returns health indicators from Tendermint.
func GetHealth(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		// the Health function returns a null object, so if it doesn't error we're good
		_, err = node.Health()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Could not fetch ndau node health: %v", err), http.StatusInternalServerError))
			return
		}

		chnode, err := ws.Node(cf.ChaosAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
			return
		}

		_, err = chnode.Health()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Could not fetch chaos node health: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(HealthResponse{HealthStatus{"Ok"}, HealthStatus{"Ok"}}))
	}
}

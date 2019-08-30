package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
)

// HealthStatus gives us the ability to add more status information later without messing up clients
type HealthStatus struct {
	Status string
}

// HealthResponse is the response from the /health endpoint.
type HealthResponse struct {
	Ndau HealthStatus
}

// GetHealth returns health indicators from Tendermint.
func GetHealth(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// the Health function returns a null object, so if it doesn't error we're good
		_, err := cf.Node.Health()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Could not fetch ndau node health: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(HealthResponse{HealthStatus{"Ok"}}))
	}
}

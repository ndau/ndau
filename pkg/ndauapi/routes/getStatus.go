package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// GetStatus returns this node's current status.
func GetStatus(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		status, err := tool.Info(node)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError(fmt.Sprintf("could not fetch status: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.Response{Sts: http.StatusOK, Bd: status})
	}
}

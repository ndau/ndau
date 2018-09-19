package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// GetGenesis returns the genesis doc from tendermint.
func GetGenesis(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		genesis, err := node.Genesis()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError(fmt.Sprintf("Could not fetch genesis: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.Response{Sts: http.StatusOK, Bd: genesis})
	}
}

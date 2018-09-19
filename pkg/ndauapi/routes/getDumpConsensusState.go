package routes

import (
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// GetDumpConsensusState returns the current consensus state.
func GetDumpConsensusState(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		consensusState, err := node.DumpConsensusState()
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError(fmt.Sprintf("Could not fetch consensus state: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.Response{Sts: http.StatusOK, Bd: consensusState})
	}
}

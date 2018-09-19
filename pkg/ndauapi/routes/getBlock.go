package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// GetBlock returns a block at a specified height.
func GetBlock(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		hp := r.URL.Query().Get("height")
		if hp == "" {
			reqres.RespondJSON(w, reqres.NewError("height parameter required", http.StatusBadRequest))
			return
		}

		height, err := strconv.ParseInt(hp, 10, 64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("height must be a valid number", http.StatusBadRequest))
			return
		}
		if height < 1 {
			reqres.RespondJSON(w, reqres.NewError("height must be greater than 0", http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("could not get node client", http.StatusInternalServerError))
			return
		}
		block, err := node.Block(&height)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError(fmt.Sprintf("could not get block: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.Response{Sts: http.StatusOK, Bd: block})
	}
}

package routes

import (
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getSystemVars(nodeAddr string, vars ...string) (map[string][]byte, error) {
	// first find a node to talk to
	node, err := ws.Node(nodeAddr)
	if err != nil {
		return nil, errors.Wrap(err, "getSystemVars")
	}
	sv, _, err := tool.Sysvars(node, vars...)
	return sv, err
}

// HandleSystemAll retrieves all the system keys at the current block height.
func HandleSystemAll(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := getSystemVars(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("reading chaos system variables", err, http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(values))
	}
}

// HandleSystemHistory returns the history of a given system variable.
func HandleSystemHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sysvar := bone.GetValue(r, "sysvar")
		if sysvar == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("sysvar parameter required", http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error getting node: %s", err), http.StatusInternalServerError))
			return
		}

		pageIndex, pageSize, errMsg, err := getPagingParams(r)
		if errMsg != "" {
			reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
			return
		}

		result, _, err := tool.SysvarHistory(node, sysvar, pageIndex, pageSize)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching key history: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

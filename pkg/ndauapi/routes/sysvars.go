package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/json2msgp"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getSystemVars(nodeAddr string, vars ...string) (map[string]interface{}, error) {
	// first find a node to talk to
	node, err := ws.Node(nodeAddr)
	if err != nil {
		return nil, errors.Wrap(err, "getSystemVars: get node")
	}
	sv, _, err := tool.Sysvars(node, vars...)
	if err != nil {
		return nil, errors.Wrap(err, "getSystemVars: fetch")
	}
	jsv, err := tool.SysvarsAsJSON(sv)
	if err != nil {
		return nil, errors.Wrap(err, "getSystemVars: convert msgp -> json")
	}
	return jsv, err
}

// HandleSystemAll retrieves all the system keys at the current block height.
func HandleSystemAll(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := getSystemVars(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("reading system variables", err, http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(values))
	}
}

// HandleSystemGet retrieves a comma-separated list of system keys at the current block height.
func HandleSystemGet(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sysvars := bone.GetValue(r, "sysvars")
		if sysvars == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("sysvars parameter required", http.StatusBadRequest))
			return
		}

		values, err := getSystemVars(cf.NodeAddress, strings.Split(sysvars, ",")...)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("reading system variables", err, http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(values))
	}
}

// HandleSystemSet constructs and returns an unsigned SetSysvar transaction
//
// This is a convenience intended to simplify sysvar handling, so that humans
// don't always need to deal with the internal msgpack encoding.
func HandleSystemSet(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sysvar := bone.GetValue(r, "sysvar")
		if sysvar == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("sysvar parameter required", http.StatusBadRequest))
			return
		}

		mdatabuf := new(bytes.Buffer)
		// Using nil for type hints relies on all numeric types in system vars being int64.
		err := json2msgp.ConvertStream(r.Body, mdatabuf, nil)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("converting input data", err, http.StatusBadRequest))
			return
		}

		ssv := ndau.SetSysvar{
			Name:  sysvar,
			Value: mdatabuf.Bytes(),
		}

		reqres.RespondJSON(w, reqres.OKResponse(ssv))
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
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching sysvar history: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

package routes

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-zoo/bone"
	"github.com/ndau/json2msgp"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getSystemVars(node cfg.TMClient, vars ...string) (map[string]interface{}, error) {
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
		values, err := getSystemVars(cf.Node)
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

		values, err := getSystemVars(cf.Node, strings.Split(sysvars, ",")...)
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

		limit, afters, err := getPagingParams(r, 1000)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("paging parm error", err, http.StatusBadRequest))
			return
		}

		after := uint64(0)
		if afters != "" {
			after, err = strconv.ParseUint(afters, 10, 64)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("parsing 'after'", err, http.StatusBadRequest))
				return
			}
		}

		result, _, err := tool.SysvarHistory(cf.Node, sysvar, after, limit)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching sysvar history: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

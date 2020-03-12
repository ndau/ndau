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
	"fmt"
	"net/http"
	"time"

	"github.com/go-zoo/bone"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/p2p"
)

// Time to wait for tendermint to respond.
const defaultTendermintTimeout = 5 * time.Second

// GetNode returns a list of nodes, including this one.
func GetNode(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// validate request
		nodeID := bone.GetValue(r, "id")
		if nodeID == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("node ID cannot be empty", http.StatusInternalServerError))
			return
		}

		// make request and get channels
		ch := tool.Nodes(cf.Node)
		var nodes []p2p.DefaultNodeInfo

		for {
			select {
			case nr, open := <-ch:
				// check error first
				if nr.Err != nil {
					reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error fetching node info: %v", nr.Err), http.StatusInternalServerError))
					return
				}

				// add nodes to response
				nodes = append(nodes, nr.Nodes...)
				if !open { // send response when channel closed
					for i := range nodes {
						if string(nodes[i].ID()) == nodeID {
							reqres.RespondJSON(w, reqres.OKResponse(nodes[i]))
							return
						}
					}
					// if not found
					reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not find node: %v", nodeID), http.StatusNotFound))
					return
				}
			case <-time.After(defaultTendermintTimeout):
				logrus.Warn("Timeout fetching cf.Node info.")
				reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node info", http.StatusInternalServerError))
				return
			}
		}
	}
}

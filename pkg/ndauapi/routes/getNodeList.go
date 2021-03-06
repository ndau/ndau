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
	"sort"
	"time"

	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/p2p"
)

// ResultNodeList represents a list of nodes.
type ResultNodeList struct {
	Nodes []p2p.DefaultNodeInfo `json:"nodes"`
}

// NodeInfo is used for sorting node pairs.
type NodeInfo struct {
	Moniker   string `json:"moniker"`
	NdauIndex int    `json:"ndau_index"`
}

// GetNodeList returns a list of nodes, including this one.
func GetNodeList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ndauNodes := getNodes(cf.Node, w, r)
		if ndauNodes != nil {
			// Monikers match between chaos and ndau, so we should be able iterate the two slices
			// in parallel.  However, for robustness, we create a map from moniker to node pair,
			// in case there is a mismatch between the chaos and ndau nodes.
			monikerMap := make(map[string]*NodeInfo)

			for i, node := range ndauNodes {
				moniker := node.Moniker
				if info, ok := monikerMap[moniker]; ok {
					info.NdauIndex = i
					monikerMap[moniker] = info
				} else {
					monikerMap[moniker] = &NodeInfo{Moniker: moniker, NdauIndex: i}
				}
			}

			// Fill a slice and sort by moniker.
			infoSlice := []*NodeInfo{}
			for _, info := range monikerMap {
				infoSlice = append(infoSlice, info)
			}
			sort.Slice(infoSlice, func(i, j int) bool {
				return infoSlice[i].Moniker < infoSlice[j].Moniker
			})

			// Convert to the desired response type.
			rnl := ResultNodeList{[]p2p.DefaultNodeInfo{}}
			for _, info := range infoSlice {
				if info.NdauIndex >= 0 {
					rnl.Nodes = append(rnl.Nodes, ndauNodes[info.NdauIndex])
				}
			}

			reqres.RespondJSON(w, reqres.OKResponse(rnl))
		}
	}
}

func getNodes(
	node cfg.TMClient,
	w http.ResponseWriter,
	r *http.Request,
) []p2p.DefaultNodeInfo {
	nodeCh := tool.Nodes(node)
	var nodes []p2p.DefaultNodeInfo

	for {
		select {
		case nr, open := <-nodeCh:
			// check error first
			if nr.Err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error fetching node info: %v", nr.Err), http.StatusInternalServerError))
				return nil
			}

			// add nodes to response
			nodes = append(nodes, nr.Nodes...)
			if !open { // send response when channel closed
				return nodes
			}
		case <-time.After(defaultTendermintTimeout):
			logrus.Warn("Timed out fetching cf.Node list.")
			reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node list", http.StatusInternalServerError))
			return nil
		}
	}
}

// GetRegisteredNodes gets the list of registered nodes in the network
func GetRegisteredNodes(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes, err := tool.RegisteredNodes(cf.Node)

		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("getting registered nodes from ndau tool", err, http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(nodes))
	}
}

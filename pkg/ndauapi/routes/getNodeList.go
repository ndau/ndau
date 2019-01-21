package routes

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tendermint/tendermint/p2p"
)

// ResultNodePair represents a chaos-ndau node pair.
type ResultNodePair struct {
	ChaosNode p2p.NodeInfo `json:"chaos"`
	NdauNode  p2p.NodeInfo `json:"ndau"`
}

// ResultNodeList represents a list of nodes.
type ResultNodeList struct {
	Nodes []ResultNodePair `json:"nodes"`
}

// NodePairInfo is used for sorting node pairs.
type NodePairInfo struct {
	Moniker    string
	ChaosIndex int
	NdauIndex  int
}

// GetNodeList returns a list of nodes, including this one.
func GetNodeList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chaosNodes := getNodes(cf.ChaosAddress, w, r)
		ndauNodes := getNodes(cf.NodeAddress, w, r)
		if chaosNodes != nil && ndauNodes != nil {
			// Monikers match between chaos and ndau, so we should be able iterate the two slices
			// in parallel.  However, for robustness, we create a map from moniker to node pair,
			// in case there is a mismatch between the chaos and ndau nodes.
			monikerMap := make(map[string]*NodePairInfo)

			// Loop over both slices, in case one is shorter than the other.
			for i, node := range chaosNodes {
				moniker := node.Moniker
				monikerMap[moniker] = &NodePairInfo{Moniker: moniker, ChaosIndex: i}
			}
			for i, node := range ndauNodes {
				moniker := node.Moniker
				if info, ok := monikerMap[moniker]; ok {
					info.NdauIndex = i
				} else {
					monikerMap[moniker] = &NodePairInfo{Moniker: moniker, NdauIndex: i}
				}
			}

			// Fill a slice and sort by moniker.
			infoSlice := []*NodePairInfo{}
			for _, info := range monikerMap {
				infoSlice = append(infoSlice, info)
			}
			sort.Slice(infoSlice, func(i, j int) bool {
				return infoSlice[i].Moniker < infoSlice[j].Moniker
			})

			// Convert to the desired response type.
			rnl := ResultNodeList{[]ResultNodePair{}}
			for _, info := range infoSlice {
				rnp := ResultNodePair{}
				if info.ChaosIndex >= 0 {
					rnp.ChaosNode = chaosNodes[info.ChaosIndex]
				}
				if info.NdauIndex >= 0 {
					rnp.NdauNode = ndauNodes[info.NdauIndex]
				}
				rnl.Nodes = append(rnl.Nodes, rnp)
			}

			reqres.RespondJSON(w, reqres.OKResponse(rnl))
		}
	}
}

func getNodes(
	nodeAddress string,
	w http.ResponseWriter,
	r *http.Request,
) []p2p.NodeInfo {
	// get node
	node, err := ws.Node(nodeAddress)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error creating node: %v", err), http.StatusInternalServerError))
		return nil
	}

	nodeCh := tool.Nodes(node)
	var nodes []p2p.NodeInfo

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
			logrus.Warn("Timed out fetching node list.")
			reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node list", http.StatusInternalServerError))
			return nil
		}
	}
}

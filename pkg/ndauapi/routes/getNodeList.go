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

// GetNodeList returns a list of nodes, including this one.
func GetNodeList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chaosNodes := getNodes(cf.ChaosAddress, w, r)
		ndauNodes := getNodes(cf.NodeAddress, w, r)
		if chaosNodes != nil && ndauNodes != nil {
			rnl := ResultNodeList{[]ResultNodePair{}}

			// Since the monikers match between chaos and ndau, we can iterate in parallel.
			chaosNodeCount := len(chaosNodes)
			ndauNodeCount := len(ndauNodes)
			maxLength := chaosNodeCount
			if ndauNodeCount > maxLength {
				maxLength = ndauNodeCount
			}
			for i := 0; i < maxLength; i++ {
				var chaosNode p2p.NodeInfo
				var ndauNode p2p.NodeInfo

				// Check lengths in case they all didn't come back from the query.
				if i < chaosNodeCount {
					chaosNode = chaosNodes[i]
				}
				if i < ndauNodeCount {
					ndauNode = ndauNodes[i]
				}

				rnp := ResultNodePair{
					ChaosNode: chaosNode,
					NdauNode:  ndauNode,
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
				sort.Slice(nodes, func(i, j int) bool {
					return string(nodes[i].Moniker) < string(nodes[j].Moniker)
				})
				return nodes
			}
		case <-time.After(defaultTendermintTimeout):
			logrus.Warn("Timed out fetching node list.")
			reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node list", http.StatusInternalServerError))
			return nil
		}
	}
}

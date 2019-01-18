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
	Moniker string
	Pair    ResultNodePair
}

// NodeResponse represents a response from the getNodes call.
type NodeResponse struct {
	Err  *reqres.APIError
	Node p2p.NodeInfo
}

// Handle the given node response.  Return true if it contained an error.
func (nr NodeResponse) handleNodeResponse(
	forChaos bool,
	monikerMap map[string]*NodePairInfo,
	w http.ResponseWriter,
) bool {
	if nr.Err != nil {
		reqres.RespondJSON(w, *nr.Err)
		return true
	}

	var info *NodePairInfo
	var ok bool
	moniker := nr.Node.Moniker
	if info, ok = monikerMap[moniker]; !ok {
		info = &NodePairInfo{Moniker: moniker, Pair: ResultNodePair{}}
		monikerMap[moniker] = info
	}
	if forChaos {
		info.Pair.ChaosNode = nr.Node
	} else {
		info.Pair.NdauNode = nr.Node
	}

	return false
}

// GetNodeList returns a list of nodes, including this one.
func GetNodeList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// We use this to pair up chaos and ndau nodes together.  They each use the same monikers.
		monikerMap := make(map[string]*NodePairInfo)

		// Fill the moniker map with nodes received through chaos and ndau channels.
		chaosCh := make(chan NodeResponse)
		ndauCh := make(chan NodeResponse)
		go getNodes(cf.ChaosAddress, chaosCh)
		go getNodes(cf.NodeAddress, ndauCh)

		// Have to use separate for-select constructs for a couple reasons:
		// 1. If one channel closes before the other, that channel will return zero values.
		// 2. If we don't close the channel, we need a nr.Done flag and then the monikerMap
		//    doesn't get filled consistently.  Seemed like a threading issue.
		var nr NodeResponse
		for open := true; open; {
			select {
			case nr, open = <-chaosCh:
				if open && nr.handleNodeResponse(true, monikerMap, w) {
					return
				}
			}
		}
		for open := true; open; {
			select {
			case nr, open = <-ndauCh:
				if open && nr.handleNodeResponse(false, monikerMap, w) {
					return
				}
			}
		}

		// Fill a slice and sort by moniker.
		var infoSlice []*NodePairInfo
		for _, info := range monikerMap {
			infoSlice = append(infoSlice, info)
		}
		sort.Slice(infoSlice, func(i, j int) bool {
			return infoSlice[i].Moniker < infoSlice[j].Moniker
		})

		// Convert to the desired response type now that the node pairs have been sorted.
		rnl := ResultNodeList{[]ResultNodePair{}}
		for _, info := range infoSlice {
			rnl.Nodes = append(rnl.Nodes, info.Pair)
		}

		reqres.RespondJSON(w, reqres.OKResponse(rnl))
	}
}

func getNodes(nodeAddress string, ch chan NodeResponse) {
	// close the channel when we're done
	defer close(ch)

	// get node
	node, err := ws.Node(nodeAddress)
	if err != nil {
		apiErr := reqres.NewAPIError(
			fmt.Sprintf("error creating node: %v", err),
			http.StatusInternalServerError)
		ch <- NodeResponse{Err: &apiErr}
	} else {
		nodeCh := tool.Nodes(node)
		for open := true; open; {
			var nr tool.NodeResponse
			select {
			case nr, open = <-nodeCh:
				// check error first
				if nr.Err != nil {
					apiErr := reqres.NewAPIError(
						fmt.Sprintf("error fetching node info: %v", nr.Err),
						http.StatusInternalServerError)
					ch <- NodeResponse{Err: &apiErr}
					return
				}

				// feed the nodes to the given channel
				for _, node := range nr.Nodes {
					ch <- NodeResponse{Node: node}
				}
			case <-time.After(defaultTendermintTimeout):
				logrus.Warn("Timed out fetching node list.")
				apiErr := reqres.NewAPIError(
					"timed out fetching node list",
					http.StatusInternalServerError)
				ch <- NodeResponse{Err: &apiErr}
				return
			}
		}
	}
}

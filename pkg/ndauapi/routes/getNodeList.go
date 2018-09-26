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

// ResultNodeList represents a list of nodes.
type ResultNodeList struct {
	Nodes []p2p.NodeInfo `json:"nodes"`
}

// GetNodeList returns a list of nodes, including this one.
func GetNodeList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get node
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error creating node: %v", err), http.StatusInternalServerError))
			return
		}

		nodeCh := tool.Nodes(node)
		resp := ResultNodeList{Nodes: []p2p.NodeInfo{}}

		for {
			select {
			case nr, open := <-nodeCh:
				// check error first
				if nr.Err != nil {
					reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error fetching node info: %v", nr.Err), http.StatusInternalServerError))
					return
				}

				// add nodes to response
				resp.Nodes = append(resp.Nodes, nr.Nodes...)
				if !open { // send response when channel closed
					sort.Slice(resp.Nodes, func(i, j int) bool {
						return string(resp.Nodes[i].ID) < string(resp.Nodes[j].ID)
					})
					reqres.RespondJSON(w, reqres.OKResponse(resp))
					return
				}
			case <-time.After(defaultTendermintTimeout):
				logrus.Warn("Timed out fetching node list.")
				reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node list", http.StatusInternalServerError))
				return

			}
		}
	}
}

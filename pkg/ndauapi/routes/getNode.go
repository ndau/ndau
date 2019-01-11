package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tendermint/tendermint/p2p"

	"github.com/go-zoo/bone"
	"github.com/sirupsen/logrus"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
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

		// get node
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error creating node: %v", err), http.StatusInternalServerError))
			return
		}

		// make request and get channels
		ch := tool.Nodes(node)
		var nodes []p2p.NodeInfo

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
				}
			case <-time.After(defaultTendermintTimeout):
				logrus.Warn("Timeout fetching node info.")
				reqres.RespondJSON(w, reqres.NewAPIError("timed out fetching node info", http.StatusInternalServerError))
				return

			}
		}
	}
}

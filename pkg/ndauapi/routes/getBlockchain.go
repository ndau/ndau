package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
)

// BlockchainRequest represents a blockchain request.
type BlockchainRequest struct {
	end   int64
	start int64
}

// MaximumRange is the maximum amount of blocks able to be returned.
const MaximumRange = 100

// StartKey represents the parameter key of the starting block height.
const StartKey = "start"

// EndKey represents the parameter key of the ending block height.
const EndKey = "end"

func processBlockchainRequest(r *http.Request) (BlockchainRequest, error) {

	var req BlockchainRequest

	keys := []string{StartKey, EndKey}
	vals := [2]int64{}
	for i, k := range keys {
		p := r.URL.Query().Get(k)
		if p == "" {
			return req, fmt.Errorf("%s parameter required", k)
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return req, fmt.Errorf("%s is not a valid number: %v", k, err)
		}
		if v < 1 {
			return req, fmt.Errorf("%s must be higher than 1", k)
		}
		vals[i] = v
	}
	start := vals[0]
	end := vals[1]

	if start > end {
		return req, fmt.Errorf("%s must be higher than %s", EndKey, StartKey)
	}
	if end-start > MaximumRange {
		return req, fmt.Errorf("%v range is larger than %v", end-start, MaximumRange)
	}

	return BlockchainRequest{
		start: start,
		end:   end,
	}, nil
}

// GetBlockchain returns the blockchain within a range.
func GetBlockchain(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := processBlockchainRequest(r)
		if err != nil {
			// Anything that errors from here is going to be a bad request.
			reqres.RespondJSON(w, reqres.NewError(err.Error(), http.StatusBadRequest))
			return
		}
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError("Could not get a node.", http.StatusInternalServerError))
			return
		}
		block, err := node.BlockchainInfo(req.start, req.end)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.Response{Sts: http.StatusOK, Bd: block})
	}
}

package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BlockchainRequest represents a blockchain request.
type BlockchainRequest struct {
	first   int64
	last    int64
	noempty bool
}

// BlockchainDateRequest represents a blockchain date request.
type BlockchainDateRequest struct {
	first   time.Time
	last    time.Time
	noempty bool
}

// MaximumRange is the maximum amount of blocks able to be returned.
const MaximumRange = 100

func processBlockchainRequest(r *http.Request) (BlockchainRequest, error) {

	var req BlockchainRequest

	keys := []string{"first", "last"}
	vals := [2]int64{}
	for i, k := range keys {
		p := bone.GetValue(r, k)
		if p == "" {
			return req, fmt.Errorf("%s parameter required", k)
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return req, fmt.Errorf("%s is not a valid number: %v", k, err)
		}
		if v < 1 {
			return req, fmt.Errorf("%s must be at least 1", k)
		}
		vals[i] = v
	}
	first := vals[0]
	last := vals[1]

	if first > last {
		return req, fmt.Errorf("%s cannot be higher than %s", keys[0], keys[1])
	}

	if last-first > MaximumRange {
		return req, fmt.Errorf("%v range is larger than %v", last-first, MaximumRange)
	}

	noempty := (r.URL.Query().Get("noempty") != "")

	return BlockchainRequest{
		first:   first,
		last:    last,
		noempty: noempty,
	}, nil
}

func filterBlockchainInfo(input *rpctypes.ResultBlockchainInfo, filter func(*tmtypes.BlockMeta) bool) *rpctypes.ResultBlockchainInfo {
	ret := &rpctypes.ResultBlockchainInfo{
		LastHeight: input.LastHeight,
		BlockMetas: make([]*tmtypes.BlockMeta, 0),
	}

	for _, p := range input.BlockMetas {
		if filter(p) {
			ret.BlockMetas = append(ret.BlockMetas, p)
		}
	}
	return ret
}

func noFilter(p *tmtypes.BlockMeta) bool {
	return true
}

func nonemptyFilter(p *tmtypes.BlockMeta) bool {
	return p.Header.NumTxs > 0
}

func hasHashOf(hash string) func(*tmtypes.BlockMeta) bool {
	return func(p *tmtypes.BlockMeta) bool {
		return p.BlockID.Hash.String() == hash
	}
}

func getCurrentBlockHeight(cf cfg.Cfg) (int64, error) {
	node, err := ws.Node(cf.NodeAddress)
	if err != nil {
		return 0, errors.New("could not get node client")
	}
	block, err := node.Block(nil)
	if err != nil {
		return 0, errors.New("could not get block")
	}
	return block.Block.Height, nil
}

func getBlocksMatching(node *client.HTTP, first, last int64, filter func(*tmtypes.BlockMeta) bool) (*rpctypes.ResultBlockchainInfo, error) {
	blocks, err := node.BlockchainInfo(first, last)
	if err != nil {
		return nil, err
	}

	return filterBlockchainInfo(blocks, filter), nil
}

func handleBlockRange(w http.ResponseWriter, r *http.Request, nodeAddress string) {
	reqdata, err := processBlockchainRequest(r)
	if err != nil {
		// Anything that errors from here is going to be a bad request.
		reqres.RespondJSON(w, reqres.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}
	node, err := ws.Node(nodeAddress)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
		return
	}

	f := noFilter
	if reqdata.noempty {
		f = nonemptyFilter
	}
	blocks, err := getBlocksMatching(node, reqdata.first, reqdata.last, f)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
		return
	}

	reqres.RespondJSON(w, reqres.OKResponse(blocks))
}

func handleBlockDateRange(w http.ResponseWriter, r *http.Request, nodeAddress string) {
	first := bone.GetValue(r, "first")
	last := bone.GetValue(r, "last")
	noempty := r.URL.Query().Get("noempty") != ""

	if first == "" {
		reqres.RespondJSON(w, reqres.NewAPIError("first parameter required.", http.StatusBadRequest))
		return
	}
	if last == "" {
		reqres.RespondJSON(w, reqres.NewAPIError("last parameter required.", http.StatusBadRequest))
		return
	}

	var err error
	_, err = time.Parse(time.RFC3339, first)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("first is not a valid timestamp", http.StatusBadRequest))
		return
	}
	_, err = time.Parse(time.RFC3339, last)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("last is not a valid timestamp", http.StatusBadRequest))
		return
	}

	node, err := ws.Node(nodeAddress)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
		return
	}

	firstHeight, lastHeight, err := tool.SearchDateRange(node, first, last)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address history: %s", err), http.StatusInternalServerError))
		return
	}

	// If the search returned zero for the last height, it means the results are empty.
	if lastHeight == 0 {
		// Successful (empty) results.
		reqres.RespondJSON(w, reqres.OKResponse(nil))
		return
	}

	// The last height is exclusive.  Otherwise the result would include the first block of the
	// next day.  The user code uses the returned range to do an inclusive block height lookup.
	// This converts it to inclusive, which is what getBlocksMatching() expects.
	lastHeight--

	// Tendermint doesn't use block height 0.
	if firstHeight == 0 {
		firstHeight = 1
	}

	// Test for degenerate results.
	if firstHeight > lastHeight {
		// Successful (empty) results.
		reqres.RespondJSON(w, reqres.OKResponse(nil))
		return
	}

	f := noFilter
	if noempty {
		f = nonemptyFilter
	}
	blocks, err := getBlocksMatching(node, int64(firstHeight), int64(lastHeight), f)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
		return
	}

	reqres.RespondJSON(w, reqres.OKResponse(blocks))
}

// HandleBlockRange handles requests for a range of blocks
func HandleBlockRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockRange(w, r, cf.NodeAddress)
	}
}

// HandleChaosBlockRange handles requests for a range of blocks
func HandleChaosBlockRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockRange(w, r, cf.ChaosAddress)
	}
}

// HandleBlockDateRange handles requests for a range of blocks
func HandleBlockDateRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockDateRange(w, r, cf.NodeAddress)
	}
}

// HandleChaosBlockDateRange handles requests for a range of blocks
func HandleChaosBlockDateRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockDateRange(w, r, cf.ChaosAddress)
	}
}

// HandleBlockHeight returns data for a single block; if height is 0, it's the current block
func HandleBlockHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pheight *int64
		hp := bone.GetValue(r, "height")
		if hp != "" {
			height, err := strconv.ParseInt(hp, 10, 64)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("height must be a valid number", err, http.StatusBadRequest))
				return
			}
			if height < 1 {
				reqres.RespondJSON(w, reqres.NewAPIError("height must be greater than 0", http.StatusBadRequest))
				return
			}
			pheight = &height
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not get node client", http.StatusInternalServerError))
			return
		}
		block, err := node.Block(pheight)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get block: %v", err), http.StatusBadRequest))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(block))
	}
}

// HandleBlockHash delivers a block matching a hash
func HandleBlockHash(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		blockhash := bone.GetValue(r, "blockhash")
		if blockhash == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("blockhash parameter required", http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not get node client", http.StatusInternalServerError))
			return
		}

		// Prepare search params.
		params := search.QueryParams{
			Command: search.HeightByBlockHashCommand,
			Hash:    blockhash, // Hex digits are query-escaped by default.
		}
		paramsBuf := &bytes.Buffer{}
		json.NewEncoder(paramsBuf).Encode(params)
		paramsString := paramsBuf.String()

		searchValue, err := tool.GetSearchResults(node, paramsString)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get search results: %v", err), http.StatusInternalServerError))
			return
		}

		blockheight, err := strconv.ParseInt(searchValue, 10, 64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not parse search results: %v", err), http.StatusInternalServerError))
			return
		}

		if blockheight <= 0 {
			// The search was valid, but there were no results.
			reqres.RespondJSON(w, reqres.OKResponse(nil))
			return
		}

		block, err := node.Block(&blockheight)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get block: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(block))
	}
}

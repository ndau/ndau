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
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
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
	// See tendermint/rpc/core/blocks.go:BlockchainInfo() for where this constant comes from.
	const pageSize int64 = 20

	// We'll build up the result, one page of blocks at a time.
	var blocks *rpctypes.ResultBlockchainInfo

	for lastIndex := last; lastIndex >= first; lastIndex -= pageSize {
		// The indexes are inclusive, hence the +1.
		firstIndex := lastIndex - pageSize + 1
		if firstIndex < first {
			firstIndex = first
		}

		blocksPage, err := node.BlockchainInfo(firstIndex, lastIndex)
		if err != nil {
			return nil, err
		}

		if blocks == nil {
			blocks = blocksPage
		} else {
			blocks.BlockMetas = append(blocks.BlockMetas, blocksPage.BlockMetas...)
		}
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

	firstTime, err := time.Parse(time.RFC3339, first)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("first is not a valid timestamp", http.StatusBadRequest))
		return
	}
	lastTime, err := time.Parse(time.RFC3339, last)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("last is not a valid timestamp", http.StatusBadRequest))
		return
	}

	pageIndex, pageSize, errMsg, err := getPagingParams(r)
	if errMsg != "" {
		reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
		return
	}

	// We sometimes support negative page index to mean "page backwards", but not here.
	// Also, the page size has already been asserted to be positive and not exceeding the max.
	if pageIndex < 0 {
		errMsg = "pagesize must be non-negative"
		reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
		return
	}

	node, err := ws.Node(nodeAddress)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
		return
	}

	firstBlockHeight := int64(1)
	firstBlock, err := node.Block(&firstBlockHeight)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get first block: %v", err), http.StatusBadRequest))
		return
	}

	lastBlock, err := node.Block(nil)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get last block: %v", err), http.StatusBadRequest))
		return
	}

	// Make sure the range is within the existing block range or the dates won't be indexed.
	firstBlockTime := firstBlock.Block.Header.Time
	lastBlockTime := lastBlock.Block.Header.Time

	if firstTime.Before(firstBlockTime) {
		// We don't index anything before the first block, clip the first time to its time.
		firstTime = firstBlockTime
		first = firstTime.Format(time.RFC3339)
	}

	if firstTime.After(lastBlockTime) {
		// Nothing is after the last block, return zero results.
		reqres.RespondJSON(w, reqres.OKResponse(nil))
		return
	}

	// Last block time is an exclusive timestamp param, so we check on-or-before the first time.
	if !lastTime.After(firstTime) {
		// Degenerate range means empty search results.
		reqres.RespondJSON(w, reqres.OKResponse(nil))
		return
	}

	// We don't check for equality here, since if the time were equal to the last block time, we'd
	// want to skip inclusion of the last block in the results.  However, since we don't index
	// every timestamp (we use a fraction of a day granularity), this is "overly correct".  But
	// there is value in this code not knowing about the underlying granularity constraints.
	if lastTime.After(lastBlockTime) {
		// Use the empty string to mean "return everything up to the newest block".
		last = ""
	}

	firstHeight, lastHeight, err := tool.SearchDateRange(node, first, last)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address history: %s", err), http.StatusInternalServerError))
		return
	}

	// getBlocksMatching() requires heights of at least 1, even if we know there to be a block at
	// height 0 (e.g. genesis system variables in the chaos chain.)
	if firstHeight == 0 {
		firstHeight = 1
	}

	// If the search returned zero for the last height, there is no chance of non-empty results.
	// We need this check so that we can decrement it safely below, as it is unsigned.
	if lastHeight == 0 {
		// Successful (empty) results.
		reqres.RespondJSON(w, reqres.OKResponse(nil))
		return
	}

	// Limit the results to the requested page.
	pagedFirstHeight := firstHeight + uint64(pageIndex*pageSize)
	if pagedFirstHeight > lastHeight {
		pagedFirstHeight = lastHeight
	}
	pagedLastHeight := pagedFirstHeight + uint64(pageSize)
	if pagedLastHeight > lastHeight {
		pagedLastHeight = lastHeight
	}
	// Replace the first and last heights with the paged subset.
	firstHeight = pagedFirstHeight
	lastHeight = pagedLastHeight

	// The last height param is exclusive.  Otherwise the result could include the first block of
	// the next day (assuming 1-day granularity under the hood).  This converts the last height to
	// inclusive, which is what getBlocksMatching() expects.
	lastHeight--

	// Test for degenerate results, both heights are inclusive at this point.
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
		block := handleBlockHeight(cf, w, r)
		if block != nil {
			reqres.RespondJSON(w, reqres.OKResponse(block))
		}
	}
}

func handleBlockHeight(cf cfg.Cfg, w http.ResponseWriter, r *http.Request) *rpctypes.ResultBlock {
	var pheight *int64
	hp := bone.GetValue(r, "height")
	if hp != "" {
		height, err := strconv.ParseInt(hp, 10, 64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("height must be a valid number", err, http.StatusBadRequest))
			return nil
		}
		if height < 1 {
			reqres.RespondJSON(w, reqres.NewAPIError("height must be greater than 0", http.StatusBadRequest))
			return nil
		}
		pheight = &height
	}

	node, err := ws.Node(cf.NodeAddress)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError("could not get node client", http.StatusInternalServerError))
		return nil
	}

	block, err := node.Block(pheight)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get block: %v", err), http.StatusBadRequest))
		return nil
	}

	return block
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

// HandleBlockTransactions delivers the transactions contained within the block specified by
// the given blockhash.
func HandleBlockTransactions(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		block := handleBlockHeight(cf, w, r)
		if block != nil {
			txHashes := []string{}

			for _, txBytes := range block.Block.Data.Txs {
				txab, err := metatx.Unmarshal(txBytes, ndau.TxIDs)
				if err != nil {
					reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not decode tx: %v", err), http.StatusInternalServerError))
					return
				}

				txHashes = append(txHashes,  metatx.Hash(txab))
			}

			reqres.RespondJSON(w, reqres.OKResponse(txHashes))
		}
	}
}

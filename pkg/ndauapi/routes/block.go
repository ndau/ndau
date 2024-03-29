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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-zoo/bone"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/ndau/ndau/pkg/ndauapi/reqres"
	"github.com/ndau/ndau/pkg/tool"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BlockchainRequest represents a blockchain request.
// first and last are the range of blocks to inspect, and limit is the maximum
// number of blocks to return (which is advisory -- the actual quantity may
// be more).
type BlockchainRequest struct {
	first   int64
	last    int64
	limit   int64
	noempty bool
}

// BlockchainDateRequest represents a blockchain date request.
type BlockchainDateRequest struct {
	first   time.Time
	last    time.Time
	noempty bool
}

// getQueryParms returns a map of query keys converted to lowercase, to the
// first version of any query. While url queries theoretically support multiple
// instances of each query parameter, we're not going to do that here.
func getQueryParms(r *http.Request) map[string]string {
	inp := bone.GetAllQueries(r)
	out := make(map[string]string)
	for k, va := range inp {
		out[strings.ToLower(k)] = va[0]
	}
	return out
}

// getFilter retrieves a filter from the spec in the query
func getFilter(qp map[string]string) (func(p *tmtypes.BlockMeta) bool, error) {
	filter := qp["filter"]
	f := noFilter
	// someday we might support more filters, hence the switch
	switch filter {
	case "notempty", "noempty", "nonempty":
		f = nonemptyFilter
	case "nofilter", "":
		// we've already set nofilter
	default:
		return f, errors.New("unknown filter " + filter)
	}
	return f, nil
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

	qp := getQueryParms(r)
	noempty := (qp["noempty"] != "")

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
	// Note - Vle: TotalTxs and NumTxs were removed from the header from tendermint v0.33.0
	// Temporary solution: ignore this filter
	// return p.Header.NumTxs > 0
	return true
}

func getCurrentBlockHeight(cf cfg.Cfg) (int64, error) {
	block, err := cf.Node.Block(nil)
	if err != nil {
		return 0, errors.New("could not get block")
	}
	return block.Block.Height, nil
}

func getBlocksMatching(node cfg.TMClient, first, last, limit int64, filter func(*tmtypes.BlockMeta) bool) (*rpctypes.ResultBlockchainInfo, error) {
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
		blocksPage = filterBlockchainInfo(blocksPage, filter)

		if blocks == nil {
			blocks = blocksPage
		} else {
			blocks.BlockMetas = append(blocks.BlockMetas, blocksPage.BlockMetas...)
		}

		// stop if we exceed the limit
		if int64(len(blocks.BlockMetas)) > limit {
			blocks.BlockMetas = blocks.BlockMetas[:limit]
			break
		}
	}

	if len(blocks.BlockMetas) > 0 {
		blocks.LastHeight = blocks.BlockMetas[len(blocks.BlockMetas)-1].Header.Height
	} else {
		blocks.LastHeight = 0
	}
	return blocks, nil
}

func handleBlockBefore(w http.ResponseWriter, r *http.Request, node cfg.TMClient) {
	befores := bone.GetValue(r, "height")
	before, err := strconv.ParseInt(befores, 10, 64)
	if err != nil || before < 1 {
		reqres.RespondJSON(w, reqres.NewFromErr("height must be a number >= 1", err, http.StatusBadRequest))
		return
	}

	qp := getQueryParms(r)
	filter, err := getFilter(qp)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("bad filter spec", err, http.StatusBadRequest))
		return
	}

	limit, _, err := getPagingParams(r, MaximumRange)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("paging error", err, http.StatusBadRequest))
		return
	}

	afters := qp["after"]
	var after int64 = 1
	if afters != "" {
		after, err = strconv.ParseInt(afters, 10, 64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("after must be a number", err, http.StatusBadRequest))
			return
		}
		if after > before || after < 1 {
			reqres.RespondJSON(w, reqres.NewAPIError("after must be between 1 and before, inclusive", http.StatusBadRequest))
			return
		}
	}

	blocks, err := getBlocksMatching(node, after, before, int64(limit), filter)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
		return
	}

	reqres.RespondJSON(w, reqres.OKResponse(blocks))
}

func handleBlockRange(w http.ResponseWriter, r *http.Request, node cfg.TMClient) {
	reqdata, err := processBlockchainRequest(r)
	if err != nil {
		// Anything that errors from here is going to be a bad request.
		reqres.RespondJSON(w, reqres.NewAPIError(err.Error(), http.StatusBadRequest))
		return
	}

	f := noFilter
	if reqdata.noempty {
		f = nonemptyFilter
	}
	blocks, err := getBlocksMatching(node, reqdata.first, reqdata.last, MaximumRange, f)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
		return
	}

	reqres.RespondJSON(w, reqres.OKResponse(blocks))
}

func handleBlockDateRange(w http.ResponseWriter, r *http.Request, node cfg.TMClient) {
	first := bone.GetValue(r, "first")
	last := bone.GetValue(r, "last")
	qp := getQueryParms(r)
	noempty := qp["noempty"] != ""

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

	limit, after, err := getPagingParams(r, MaximumRange)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewFromErr("paging error", err, http.StatusBadRequest))
		return
	}

	// for this query, "after" must be a timestamp
	// we can actually use it to further constrain the "first" timestamp; the initial query
	// can be repeated with the same timestamps, but the "after" parameter can vary
	// to page the results.
	if after != "" {
		aftertime, err := time.Parse(time.RFC3339, after)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("after must be a timestamp", err, http.StatusBadRequest))
			return
		}
		if aftertime.After(firstTime) {
			firstTime = aftertime
		}
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
	var firstBlockTime, lastBlockTime time.Time
	if firstBlock != nil && firstBlock.Block != nil {
		firstBlockTime = firstBlock.Block.Header.Time
	}
	if lastBlock != nil && lastBlock.Block != nil {
		lastBlockTime = lastBlock.Block.Header.Time
	}

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

	// Limit the results
	if firstHeight > lastHeight {
		firstHeight = lastHeight
	}
	pagedLastHeight := firstHeight + uint64(limit)
	if pagedLastHeight > lastHeight {
		pagedLastHeight = lastHeight
	}
	// Replace the last height with the paged subset.
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
	blocks, err := getBlocksMatching(node, int64(firstHeight), int64(lastHeight), int64(limit), f)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get blockchain: %v", err), http.StatusInternalServerError))
		return
	}

	reqres.RespondJSON(w, reqres.OKResponse(blocks))
}

// HandleBlockBefore handles requests for blocks on or before a given height, and
// manages filtering better than HandleBlockRange.
func HandleBlockBefore(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockBefore(w, r, cf.Node)
	}
}

// HandleBlockRange handles requests for a range of blocks
func HandleBlockRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockRange(w, r, cf.Node)
	}
}

// HandleBlockDateRange handles requests for a range of blocks
func HandleBlockDateRange(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleBlockDateRange(w, r, cf.Node)
	}
}

// HandleBlockHeight returns data for a single block; if height is 0, it's the current block
func HandleBlockHeight(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		block, apiErr := handleBlockHeight(cf, r)
		if apiErr.Status() != http.StatusOK {
			reqres.RespondJSON(w, apiErr)
		} else {
			reqres.RespondJSON(w, reqres.OKResponse(block))
		}
	}
}

func handleBlockHeight(cf cfg.Cfg, r *http.Request) (*rpctypes.ResultBlock, reqres.APIError) {
	var pheight *int64
	hp := bone.GetValue(r, "height")
	if hp != "" {
		height, err := strconv.ParseInt(hp, 10, 64)
		if err != nil {
			return nil, reqres.NewFromErr("height must be a valid number", err, http.StatusBadRequest)
		}
		if height < 1 {
			return nil, reqres.NewAPIError("height must be greater than 0", http.StatusBadRequest)
		}
		pheight = &height
	}

	block, err := cf.Node.Block(pheight)
	if err != nil {
		return nil, reqres.NewAPIError(fmt.Sprintf("could not get block: %v", err), http.StatusBadRequest)
	}

	return block, reqres.NewAPIError("", http.StatusOK)
}

// HandleBlockHash delivers a block matching a hash
func HandleBlockHash(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		blockhash := bone.GetValue(r, "blockhash")
		if blockhash == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("blockhash parameter required", http.StatusBadRequest))
			return
		}

		// Prepare search params.
		params := search.QueryParams{
			Command: search.HeightByBlockHashCommand,
			Hash:    blockhash, // Hex digits are query-escaped by default.
		}

		searchValue, err := tool.GetSearchResults(cf.Node, params)
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

		block, err := cf.Node.Block(&blockheight)
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
		block, apiErr := handleBlockHeight(cf, r)
		if apiErr.Status() != http.StatusOK {
			reqres.RespondJSON(w, apiErr)
		} else {
			txHashes := []string{}

			for _, txBytes := range block.Block.Data.Txs {
				txab, err := metatx.Unmarshal(txBytes, ndau.TxIDs)
				if err != nil {
					reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not decode tx: %v", err), http.StatusInternalServerError))
					return
				}

				txHashes = append(txHashes, metatx.Hash(txab))
			}

			reqres.RespondJSON(w, reqres.OKResponse(txHashes))
		}
	}
}

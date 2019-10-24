package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// The public API for searching the index.

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// SearchSysvarHistory returns value history for the given sysvar using an index under the hood.
// The response is sorted by ascending block height, each entry is where the key's value changed.
// Pass in 0,0 for the paging params to get the entire history.
func (search *Client) SearchSysvarHistory(
	sysvar string, afterHeight uint64, limit int,
) (khr *query.SysvarHistoryResponse, err error) {
	khr = new(query.SysvarHistoryResponse)

	searchKey := fmtSysvarKeyToValue(sysvar)

	// We'll reuse this for unmarshaling data into it.
	valueData := &ValueData{}

	err = search.Client.SScan(searchKey, func(searchValue string) error {
		err := valueData.Unmarshal(searchValue)
		if err != nil {
			return err
		}

		height := valueData.Height
		valueBase64 := valueData.ValueBase64

		value, err := base64.StdEncoding.DecodeString(valueBase64)
		if err != nil {
			return err
		}

		khr.History = append(khr.History, query.SysvarHistoricalValue{
			Height: height,
			Value:  value,
		})

		return nil
	})

	// Sort by ascending height.  Even if we had used ZAdd() with ZScan(), we'd still have to sort
	// because of the way it returns pages of results under the hood that we'd have to merge.
	sort.Slice(khr.History, func(i, j int) bool {
		return khr.History[i].Height < khr.History[j].Height
	})

	// Reduce the full results list down to the requested portion.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart := sort.Search(len(khr.History), func(n int) bool {
		return khr.History[n].Height > afterHeight
	})
	khr.History = khr.History[offsetStart:]
	if limit > 0 && len(khr.History) > limit {
		khr.History = khr.History[:limit]
	}

	return khr, err
}

// SearchBlockHash returns the height of the given block hash.
// Returns 0 and no error if the given block hash was not found in the index.
func (search *Client) SearchBlockHash(blockHash string) (uint64, error) {
	searchKey := fmtBlockHashToHeight(blockHash)

	searchValue, err := search.Client.Get(searchKey)
	if err != nil {
		return 0, err
	}

	if searchValue == "" {
		// Empty search results.  No error.
		return 0, nil
	}

	return strconv.ParseUint(searchValue, 10, 64)
}

// SearchTxHash returns tx data for the given tx hash.
func (search *Client) SearchTxHash(txHash string) (TxValueData, error) {
	valueData := TxValueData{}
	searchKey := fmtTxHashToHeight(txHash)

	searchValue, err := search.Client.Get(searchKey)
	if err != nil {
		return valueData, err
	}

	if searchValue == "" {
		// Empty search results.  No error.
		return valueData, err
	}

	err = valueData.Unmarshal(searchValue)
	return valueData, err
}

// SearchTxTypes returns tx data for a range of transactions on or before the given tx hash.
// If txHashOrHeight is "", this will return the latest page of transactions from the blockchain.
// txHashOrHeight can be a block height.  Transactions in and before that block are returned.
// If txTypes is empty, this will return zero results.
// If limit is non-positive, this will return results as if the page size is infinite.
func (search *Client) SearchTxTypes(txHashOrHeight string, txTypes []string, limit int) (TxListValueData, error) {
	listValueData := TxListValueData{}

	// Treat negative as zero (no limit).
	if limit < 0 {
		limit = 0
	}

	var searchKeys []string
	if len(txTypes) == 0 {
		// No types given means all results.
		var err error
		searchKeys, err = search.Client.Keys(fmtTxTypeToHeight("*"))
		if err != nil {
			return listValueData, err
		}
	} else {
		for _, t := range txTypes {
			searchKeys = append(searchKeys, fmtTxTypeToHeight(t))
		}
	}

	// Use a unique key name for each query.
	// NOTE: We could consider using a well-defined name based on the query params, but the
	// usefulness of that is primarily around multiple blockchain explorer clients hitting the
	// same endpoint near-simultaneously.  The most common one being the first page of results.
	// However, once we have higher volume, the latest page will be ever-changing.  And, in a low
	// volume ecosystem, there are likely not going to be many simultaneous queries.  So current
	// implementation uses a short-lived union key and we DELETE it, rather than EXPIRE it.
	searchKey := fmtUnion()

	// Delete the short-lived union key from the database when we're all done.
	defer search.Client.Del(searchKey)

	// Merge sort all the requested tx type results.
	count, err := search.Client.ZUnionStore(searchKey, searchKeys)
	if err != nil || count == 0 {
		// This will return an empty list with no error if the union returned zero count.
		return listValueData, err
	}

	// Check whether we have a hash or height.
	// Tx hashes are always length 22 which is more digits than can fit in a unit64.
	// This will fail to convert to uint64 in that case and we'll assume it's a hash.
	// Same for empty string, which will be treated as "start from latest block".
	height, err := strconv.ParseUint(txHashOrHeight, 10, 64)
	if err != nil {
		height = 0
	}

	var hashes []string
	if height > 0 {
		// This is exclusive, so we add 1.  We'll get all transactions in the input block this
		// way, not just the first (tx offset 0) transaction.
		score := float64(height + 1)
		count = int64(limit)
		// If not searching all, include one more to get the tx hash for the next page.
		if count > 0 {
			count++
		}
		hashes, err = search.Client.ZRevRangeByScore(searchKey, score, count)
		if err != nil {
			return listValueData, err
		}
	} else {
		// Get the rank of the starting tx hash.  If it's not in the list, we have a bad query.
		// Use reverse rank since we want transactions in reverse chronological order.
		// Default is to start from zero (latest transaction) when an empty tx hash is given.
		var start int64
		if txHashOrHeight != "" {
			start, err = search.Client.ZRevRank(searchKey, txHashOrHeight)
			if err != nil {
				// No error if the hash is bad (not part of the results, invalid, etc).
				// Just return empty results.
				return listValueData, nil
			}
		}

		var stop int64
		if limit == 0 {
			// Zero or negative limit means "unlimited".
			stop = -1
		} else {
			// Like the start rank, the stop rank is inclusive.
			// We don't subtract one here since we want to get the next tx hash after the page.
			// If this is more than there are results, all results after the start are returned.
			stop = start + int64(limit)
		}

		// Get a page's worth of results from the union.
		hashes, err = search.Client.ZRevRange(searchKey, start, stop)
		if err != nil {
			return listValueData, err
		}
	}

	if limit == 0 || limit >= len(hashes) {
		// We're getting the last page of results, there will be no "next" tx hash.
		limit = len(hashes)
	} else {
		// The last element is used for the next tx hash and not included in our results below.
		limit = len(hashes) - 1
		listValueData.NextTxHash = hashes[limit]
	}

	// Pull out transaction data using the tx hash index, for each tx hash in the list.
	for i := 0; i < limit; i++ {
		valueData, err := search.SearchTxHash(hashes[i])
		if err != nil {
			return listValueData, err
		}

		// Sanity check.  Every search in this loop is exepected to return a valid tx result.
		if valueData.BlockHeight > 0 {
			listValueData.Txs = append(listValueData.Txs, valueData)
		}
	}

	return listValueData, nil
}

// SearchAccountHistory returns an array of block height and txoffset pairs associated with the
// given account address.
// Pass in 0, 0 for the paging params to get the entire history.
func (search *Client) SearchAccountHistory(
	addr string, afterHeight uint64, limit int,
) (ahr *AccountHistoryResponse, err error) {
	ahr = new(AccountHistoryResponse)

	searchKey := fmtAddressToHeight(addr)

	err = search.Client.SScan(searchKey, func(searchValue string) error {
		valueData := AccountTxValueData{}
		err := valueData.Unmarshal(searchValue)
		if err != nil {
			return err
		}

		ahr.Txs = append(ahr.Txs, valueData)

		return nil
	})

	// Sort by ascending height, with ascending tx offset as the secondary sort order.
	// Even if we had used ZAdd() with ZScan(), we'd still have to sort because of the way it
	// returns pages of results under the hood that we'd have to merge.
	sort.Slice(ahr.Txs, func(i, j int) bool {
		txi := &ahr.Txs[i]
		txj := &ahr.Txs[j]
		return txi.BlockHeight < txj.BlockHeight ||
			txi.BlockHeight == txj.BlockHeight && txi.TxOffset < txj.TxOffset
	})

	// Reduce the full results list down to the requested portion.  There is some wasted effort with
	// this approach, but we support the worst case, which is to return all results.  In practice,
	// getting the full list from the underlying index is fast, with tolerable sorting speed.
	offsetStart := sort.Search(len(ahr.Txs), func(n int) bool {
		return ahr.Txs[n].BlockHeight > afterHeight
	})
	ahr.Txs = ahr.Txs[offsetStart:]
	// if we need to truncate the list, notify the caller that we've done so
	if limit > 0 && len(ahr.Txs) > limit {
		ahr.Txs = ahr.Txs[:limit]
		ahr.More = true
	}

	return ahr, err
}

// BlockTime returns the timestamp for the block at a given height
// returns the zero value and no error if the block is unknown
func (search *Client) BlockTime(height uint64) (math.Timestamp, error) {
	ts, err := search.Client.Get(
		fmtHeightToTimestamp(height),
	)
	var t math.Timestamp
	if err != nil {
		return t, fmt.Errorf("getting timestamp for block %d: %s", height, err)
	}
	if ts == "" {
		return t, nil
	}
	t, err = math.ParseTimestamp(ts)
	if err != nil {
		return t, fmt.Errorf("parsing timestamp for block %d (%s): %s", height, ts, err)
	}
	return t, nil
}

// SearchMostRecentRegisterNode returns tx data for the most recent
// RegisterNode transactions for the given address.
//
// Returns a nil for TxValueData if the node has never been registered.
func (search *Client) SearchMostRecentRegisterNode(address string) (*TxValueData, error) {
	searchKeys := []string{
		fmtAddressToHeight(address),
		fmtTxTypeToHeight("RegisterNode"),
	}

	// Use a unique key name for each query.
	queryID := fmtUnion()
	// Delete the short-lived union key from the database when we're all done.
	defer search.Client.Del(queryID)

	// Merge sort all the requested results.
	count, err := search.Client.ZUnionStore(queryID, searchKeys)
	if err != nil || count == 0 {
		// This will return nil list with no error if the union successfully returned zero count.
		return nil, errors.Wrap(err, "searching redis by composite query")
	}

	// Get the rank of the starting tx hash.  If it's not in the list, we have a bad query.
	// Use reverse rank since we want transactions in reverse chronological order.
	// Default is to start from zero (latest transaction) when an empty tx hash is given.

	// Get a page's worth of results from the union.
	hashes, err := search.Client.ZRevRange(queryID, 0, -1)
	if err != nil {
		return nil, err
	}
	if len(hashes) == 0 {
		return nil, nil
	}

	// Pull out transaction data using the tx hash index, for each tx hash in the list.
	valueData, err := search.SearchTxHash(hashes[0])
	if err != nil {
		return &valueData, err
	}

	// What if redis insists on giving us a zero value?
	if valueData.BlockHeight <= 0 {
		return nil, nil
	}

	return &valueData, nil
}

// SearchMarketPrice searches for market price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (search *Client) SearchMarketPrice(params PriceQueryParams) (PriceQueryResults, error) {
	return search.searchPrice(params, marketPriceKeysetKey, marketPriceKeyFmt)
}

// SearchTargetPrice searches for target price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (search *Client) SearchTargetPrice(params PriceQueryParams) (PriceQueryResults, error) {
	return search.searchPrice(params, targetPriceKeysetKey, targetPriceKeyFmt)
}

// searchPrice searches for market price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (search *Client) searchPrice(
	params PriceQueryParams,
	key, kfmt string,
) (PriceQueryResults, error) {
	// setup search options
	var zropts redis.ZRangeBy
	if after := params.After.GetTimestamp(search); after != 0 {
		// leading paren in query causes exclusive semantics
		zropts.Min = fmt.Sprintf("(%v", float64(after))
	} else {
		zropts.Min = "-inf"
	}
	if before := params.Before.GetTimestamp(search); before != 0 {
		// leading paren in query causes exclusive semantics
		zropts.Max = fmt.Sprintf("(%v", float64(before))
	} else {
		zropts.Max = "+inf"
	}

	var iqty uint
	if params.Limit != 0 {
		// we add one so we can tell if extra elements exist
		zropts.Count = int64(params.Limit + 1)
		iqty = params.Limit
	}

	// execute query
	ks, err := search.Client.Inner().ZRangeByScore(key, zropts).Result()
	if err != nil {
		return PriceQueryResults{}, errors.Wrap(err, "querying redis")
	}
	if iqty == 0 {
		iqty = uint(len(ks))
	}

	// convert output into a nice format
	out := PriceQueryResults{
		Items: make([]PriceQueryResult, 0, iqty),
		More:  params.Limit != 0 && len(ks) > int(params.Limit),
	}
	for i, k := range ks {
		if uint(i) >= iqty {
			// don't append the extra to the output
			break
		}
		// extract height and timestamp from key
		var h uint64
		var tss string
		_, err := fmt.Sscanf(k, kfmt, &h, &tss)
		if err != nil {
			return out, errors.Wrap(
				err,
				fmt.Sprintf(
					"parsing price zset key '%s' (idx %d)",
					k,
					i,
				),
			)
		}
		// parse real timestamp
		ts, err := math.ParseTimestamp(tss)
		if err != nil {
			return out, errors.Wrap(
				err,
				fmt.Sprintf(
					"parsing zset key timestamp '%s' (idx %d)",
					tss,
					i,
				),
			)
		}
		// get price as string since we know its key
		ps, err := search.Client.Inner().Get(k).Result()
		if err != nil {
			return out, errors.Wrap(err, fmt.Sprintf("getting redis key '%s' (zset idx %d)", k, i))
		}
		// parse real price
		p, err := strconv.ParseInt(ps, 10, 64)
		if err != nil {
			return out, errors.Wrap(err, fmt.Sprintf("parsing stored price '%s' as int (zset idx %d)", ps, i))
		}
		// append this row
		out.Items = append(out.Items, PriceQueryResult{
			Price:     pricecurve.Nanocent(p),
			Height:    h,
			Timestamp: ts,
		})
	}

	return out, nil
}

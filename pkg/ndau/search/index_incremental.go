package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// Methods used for incremental indexing.

import (
	"encoding/base64"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

// OnBeginBlock resets our local cache for incrementally indexing the block at the given height.
func (search *Client) OnBeginBlock(height uint64, blockTime math.Timestamp, tmHash string) error {
	// There's only one block to consider for incremental indexing.
	search.sysvarKeyToValueData = make(map[string]*ValueData)
	search.txs = nil
	search.blockTime = blockTime
	search.blockHash = tmHash
	search.blockHeight = height
	search.nextHeight = height + 1
	return nil
}

// OnDeliverTx grabs the fields out of this transaction to index when the block is committed.
func (search *Client) OnDeliverTx(appI interface{}, tx metatx.Transactable) error {
	search.txs = append(search.txs, tx)

	app := appI.(AppIndexable)

	if indexable, ok := tx.(SysvarIndexable); ok {
		key := indexable.GetName()
		valueBase64 := base64.StdEncoding.EncodeToString(indexable.GetValue())

		searchKey := fmtSysvarKeyToValue(key)
		data, hasValue := search.sysvarKeyToValueData[searchKey]
		if hasValue {
			// Override whatever value was there before for this block.
			// We only want one k-v pair per block height in our index: the one for the latest value.
			data.ValueBase64 = valueBase64
		} else {
			search.sysvarKeyToValueData[searchKey] = &ValueData{
				Height:      search.blockHeight,
				ValueBase64: valueBase64,
			}
		}
	}

	if indexable, ok := tx.(MarketPriceIndexable); ok {
		search.marketPrice = indexable.GetMarketPrice()
	}

	if _, ok := tx.(TargetPriceIndexable); ok {
		state := app.GetState().(*backing.State)
		search.targetPrice = state.TargetPrice
	}

	return nil
}

// OnCommit indexes all the transaction data we collected since the last BeginBlock().
func (search *Client) OnCommit() error {
	_, _, err := search.index()
	if err != nil {
		return err
	}

	// We don't need to check for dupes in the new block since we filtered them out by using the
	// sysvarKeyToValueData map.  However we do need to check for dupes from values in earlier
	// blocks.
	_, _, err = search.onIndexingComplete(true)

	return err
}

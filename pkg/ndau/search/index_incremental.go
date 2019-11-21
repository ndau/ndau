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
	"context"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TODO: can we push those panics back somehow, maybe get some kind of reattempt?

// InitChain implements metaapp.Indexer
func (client *Client) InitChain(abci.RequestInitChain, abci.ResponseInitChain, metast.State) {
	// noop
}

// BeginBlock implements metaapp.Indexer
func (client *Client) BeginBlock(
	request abci.RequestBeginBlock,
	response abci.ResponseBeginBlock,
	state metast.State,
) {
	client.height = uint64(request.Header.Height)
	_, err := client.postgres.Exec(
		context.Background(),
		"INSERT INTO blocks(height, block_time) VALUES ($1, $2)",
		request.Header.Height, request.Header.Time,
	)
	if err != nil {
		panic(err)
	}
}

// DeliverTx implements metaapp.Indexer
func (client *Client) DeliverTx(
	abci.RequestDeliverTx,
	abci.ResponseDeliverTx,
	metatx.Transactable,
	metast.State,
) {
	// TODO!
}

// EndBlock implements metaapp.Indexer
func (client *Client) EndBlock(
	abci.RequestEndBlock,
	abci.ResponseEndBlock,
	metast.State,
) {
	// TODO!
}

// Commit implements metaapp.Indexer
func (client *Client) Commit(
	abci.ResponseCommit,
	metast.State,
) {
	// TODO!
}

// // OnDeliverTx grabs the fields out of this transaction to index when the block is committed.
// func (search *Client) OnDeliverTx(appI interface{}, tx metatx.Transactable) error {
// 	search.txs = append(search.txs, tx)

// 	app := appI.(AppIndexable)

// 	if indexable, ok := tx.(SysvarIndexable); ok {
// 		key := indexable.GetName()
// 		valueBase64 := base64.StdEncoding.EncodeToString(indexable.GetValue())

// 		searchKey := fmtSysvarKeyToValue(key)
// 		data, hasValue := search.sysvarKeyToValueData[searchKey]
// 		if hasValue {
// 			// Override whatever value was there before for this block.
// 			// We only want one k-v pair per block height in our index: the one for the latest value.
// 			data.ValueBase64 = valueBase64
// 		} else {
// 			search.sysvarKeyToValueData[searchKey] = &ValueData{
// 				Height:      search.blockHeight,
// 				ValueBase64: valueBase64,
// 			}
// 		}
// 	}

// 	if indexable, ok := tx.(MarketPriceIndexable); ok {
// 		search.marketPrice = indexable.GetMarketPrice()
// 	}

// 	if _, ok := tx.(TargetPriceIndexable); ok {
// 		state := app.GetState().(*backing.State)
// 		search.targetPrice = state.TargetPrice
// 	}

// 	return nil
// }

// // OnCommit indexes all the transaction data we collected since the last BeginBlock().
// func (search *Client) OnCommit() error {
// 	_, _, err := search.index()
// 	if err != nil {
// 		return err
// 	}

// 	// We don't need to check for dupes in the new block since we filtered them out by using the
// 	// sysvarKeyToValueData map.  However we do need to check for dupes from values in earlier
// 	// blocks.
// 	_, _, err = search.onIndexingComplete(true)

// 	return err
// }

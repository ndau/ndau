package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// The public API for initial indexing of the blockchain.

import (
	"fmt"
	"time"

	"github.com/attic-labs/noms/go/datas"
	"github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
)

// IndexBlockchain fills the index with data from the blockchain,
// from the head block down to just before the last block we indexed.
func (search *Client) IndexBlockchain(
	db datas.Database, ds datas.Dataset,
) (updateCount int, insertCount int, err error) {
	updateCount = 0
	insertCount = 0

	// Start fresh.  It should already be zero'd out upon entry.
	search.sysvarKeyToValueData = make(map[string]*ValueData)
	search.txs = nil
	// TODO: We really should be using block time when indexing below, but we don't store block
	// timestamps in noms.  So we must eventually write the external ndauindexer app.  Then we
	// would remove this entire function since it would replace the initial indexing below.
	search.blockTime = time.Time{}
	search.blockHash = ""
	search.blockHeight = 0
	search.nextHeight = 0

	// The height encountered on the previous iteration of the loop below.
	lastHeight := search.nextHeight

	// One more than the height we indexed to the last time we indexed the blockchain.
	// In other words, it's the height we want to index to this time.
	minHeightToIndex := search.Client.GetNextHeight()

	example := backing.State{}
	err = state.IterHistory(db, ds, &example, func(stI state.State, height uint64) error {
		// Save off the max height that we'll index up to by the end of the iteration.
		if search.nextHeight == 0 {
			// This assumes we're iterating blocks in order from the head to genesis.
			search.nextHeight = height + 1
			lastHeight = search.nextHeight
		}

		// If we've reached the last height we indexed to, we can stop here.
		if height < minHeightToIndex {
			// This assumes we're iterating blocks in order from the head to genesis.
			return state.StopIteration()
		}

		// Make sure we're iterating from the head block to genesis.
		// It's expected, however, to get multiple height-0 entries for system variables.
		// And also at height 1 for unit tests.
		if height > lastHeight || height == lastHeight && height > 1 {
			// Indexing logic relies on this, but more importantly, this indicates
			// a serious problem in the blockchain if the height increases as we
			// crawl the blockchain from the head to the genesis block.
			panic(fmt.Sprintf("Invalid height found in noms: %d >= %d", height, lastHeight))
		}
		lastHeight = height

		// The indexing code below uses this to know the current height.
		search.blockHeight = height

		// NOTE: This is currently a no-op since we didn't pull anything out of noms to put into
		// the search client struct before calling index().  We keep this here in case we do find
		// more that we want to index that also can be pulled from noms before this line.
		updCount, insCount, err := search.index()
		updateCount += updCount
		insertCount += insCount
		if err != nil {
			return err
		}

		// Index sysvar key-value history, which we pull from noms.
		st := stI.(*backing.State)
		updCount, insCount, err = search.indexState(st)
		updateCount += updCount
		insertCount += insCount
		return err
	})
	if err != nil && !state.IsStopIteration(err) {
		return updateCount, insertCount, err
	}

	// We don't need to check for dupes if this is the first initial scan (minHeightToIndex
	// == 0) since we will have just completed indexing the entire blockchain and have filtered
	// out all the dupes from adjacent (and within) blocks using sysvarKeyToValueData map.
	// We also don't need to check for dupes if we didn't index anything this time
	// (minHeightToIndex == search.nextHeight), although that check is just for completeness
	// since sysvarKeyToValueData will be empty in that case.
	checkForDupes := 0 < minHeightToIndex && minHeightToIndex < search.nextHeight
	updCount, insCount, err := search.onIndexingComplete(checkForDupes)
	updateCount += updCount
	insertCount += insCount

	return updateCount, insertCount, err
}

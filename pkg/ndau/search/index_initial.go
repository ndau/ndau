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
	"context"
	"fmt"

	"github.com/attic-labs/noms/go/datas"
	"github.com/jackc/pgx"
	"github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/pkg/errors"
)

// IndexBlockchain fills the index with data from the blockchain,
// from the head block down to just before the last block we indexed.
func (client *Client) IndexBlockchain(
	db datas.Database, ds datas.Dataset,
) (err error) {
	// what's the greatest block height we've already indexed?
	var maxHeightAlreadyIndexed uint64
	err = client.postgres.QueryRow(context.Background(), "SELECT MAX(height) FROM blocks").
		Scan(&maxHeightAlreadyIndexed)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "querying max height already indexed")
	}

	var lastHeight uint64

	example := backing.State{}
	err = state.IterHistory(db, ds, &example, func(stI state.State, height uint64) error {
		// keep track of the last height viewed
		defer func() { lastHeight = height }()
		// If we've reached the last height we indexed to, we can stop here.
		if height <= maxHeightAlreadyIndexed && maxHeightAlreadyIndexed > 0 {
			return state.StopIteration()
		}

		// Make sure we're iterating from the head block to genesis.
		// It's expected, however, to get multiple height-0 entries for system variables.
		// And also at height 1 for unit tests.
		if height > lastHeight || height == lastHeight && height > 1 {
			// Indexing logic relies on this, but more importantly, this indicates
			// a serious problem in the blockchain if the height increases as we
			// crawl the blockchain from the head to the genesis block.
			panic(fmt.Sprintf("invalid height found in noms: %d >= %d", height, lastHeight))
		}

		// NOTE: This is currently a no-op since we didn't pull anything out of noms to put into
		// the search client struct before calling index().  We keep this here in case we do find
		// more that we want to index that also can be pulled from noms before this line.
		_, _, err := client.index()
		if err != nil {
			return err
		}

		// Index sysvar key-value history, which we pull from noms.
		st := stI.(*backing.State)
		_, _, err = client.indexState(st)
		return err
	})
	if err != nil && !state.IsStopIteration(err) {
		return errors.Wrap(err, "iterating noms history")
	}

	// We don't need to check for dupes if this is the first initial scan (minHeightToIndex
	// == 0) since we will have just completed indexing the entire blockchain and have filtered
	// out all the dupes from adjacent (and within) blocks using sysvarKeyToValueData map.
	// We also don't need to check for dupes if we didn't index anything this time
	// (minHeightToIndex == search.nextHeight), although that check is just for completeness
	// since sysvarKeyToValueData will be empty in that case.
	_, _, err = client.onIndexingComplete(false)

	return
}

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
	err = client.Postgres.QueryRow(
		context.Background(),
		"SELECT COALESCE(MAX(height), 0) FROM blocks",
	).
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

		// Index sysvar key-value history, which we pull from noms.
		// this is broken out into a function mainly for ease of reading/separation
		// of concerns: the logic around here is all about iterating through noms;
		// the logic there is about actually performing appropriate indexing
		client.height = height
		return client.indexInitialSysvars(stI.(*backing.State))
	})
	// if the returned error was nil, this preserves its nility
	return errors.Wrap(err, "iterating noms history")
}

// Index all the sysvar key-value pairs in the given state at the current client.height.
func (client *Client) indexInitialSysvars(st *backing.State) (err error) {
	for key, value := range st.Sysvars {
		// note that for this initial indexing, we manually dedupe; we expect
		// to see more than one entry at height 0, and there's no point inserting
		// redundant initial sysvar data
		_, err = client.Postgres.Exec(
			context.Background(),
			"INSERT INTO systemvariables(height, key, value) "+
				"SELECT $1, $2, $3 "+
				"WHERE NOT EXISTS ("+
				"SELECT key FROM systemvariables WHERE height=$1 AND key=$2 AND value=$3"+
				")",
			client.height,
			key,
			value,
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("%s@%d", key, client.height))
		}
	}

	return
}

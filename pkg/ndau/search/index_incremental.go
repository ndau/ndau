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
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	abci "github.com/tendermint/tendermint/abci/types"
)

// TODO: can we push those panics back somehow, maybe get some kind of reattempt?

// InitChain implements metaapp.Indexer
func (client *Client) InitChain(abci.RequestInitChain, abci.ResponseInitChain, metast.State) {
	// noop, but if we ever want to keep track of validator updates, we can't
	// forget to handle this initial case returned here
}

// BeginBlock implements metaapp.Indexer
func (client *Client) BeginBlock(
	request abci.RequestBeginBlock,
	response abci.ResponseBeginBlock,
	stI metast.State,
) {
	client.height = uint64(request.Header.Height)
	client.sequence = 0
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
	request abci.RequestDeliverTx,
	response abci.ResponseDeliverTx,
	tx metatx.Transactable,
	stI metast.State,
) {
	defer func() { client.sequence++ }()
	// we can't handle errors in these calculations, and worst case, we get
	// a zero back, so... we just discard potential errors
	fee, _ := client.app.CalculateTxFeeNapu(tx)
	sib, _ := client.app.CalculateTxSIBNapu(tx)

	_, err := client.postgres.Exec(
		context.Background(),
		"INSERT INTO transactions(name, hash, height, sequence, data, fee, sib) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7)",
		metatx.NameOf(tx),
		metatx.Hash(tx),
		client.height,
		client.sequence,
		tx,
		fee,
		sib,
	)
	if err != nil {
		panic(err)
	}

	// a few other tables need the tx row.
	// In high-latency systems, it might be faster to fetch this as a subquery,
	// but we know that postgres will always be on the same physical machine as
	// the node software, so it is probably worth it to cache it ahead of time.
	var txRow uint64
	err = client.postgres.QueryRow(
		context.Background(),
		"SELECT id FROM transactions WHERE block=$1 AND sequence=$2 LIMIT 1",
		client.height,
		client.sequence,
	).Scan(&txRow)

	state := stI.(*backing.State)
	accountsAffected, err := client.app.GetAccountAddresses(tx)
	if err != nil {
		return // not worth a panic
	}
	for _, addr := range accountsAffected {
		if ad, ok := state.Accounts[addr]; ok {
			_, err = client.postgres.Exec(
				context.Background(),
				"INSERT INTO accounts(address, data, tx) "+
					"VALUES ($1, $2, $3)",
				addr, ad, txRow,
			)
			if err != nil {
				panic(err)
			}
		}
	}

	if indexable, ok := tx.(SysvarIndexable); ok {
		_, err = client.postgres.Exec(
			context.Background(),
			"INSERT INTO systemvariables(key, value, height, tx) "+
				"VALUES ($1, $2, $3, $4)",
			indexable.GetName(),
			indexable.GetValue(),
			client.height,
			txRow,
		)
		if err != nil {
			panic(err)
		}
	}

	if indexable, ok := tx.(MarketPriceIndexable); ok {
		_, err = client.postgres.Exec(
			context.Background(),
			"INSERT INTO marketprices(tx, price) VALUES ($1, $2)",
			txRow,
			indexable.GetMarketPrice(),
		)
		if err != nil {
			panic(err)
		}
	}

	if _, ok := tx.(TargetPriceIndexable); ok {
		_, err = client.postgres.Exec(
			context.Background(),
			"INSERT INTO targetprices(tx, price) VALUES ($1, $2)",
			txRow,
			state.TargetPrice,
		)
		if err != nil {
			panic(err)
		}
	}
}

// EndBlock implements metaapp.Indexer
func (client *Client) EndBlock(
	request abci.RequestEndBlock,
	response abci.ResponseEndBlock,
	stI metast.State,
) {
	// noop, but if we ever want to keep track of validator updates, this is
	// the place to do it
}

// Commit implements metaapp.Indexer
func (client *Client) Commit(
	response abci.ResponseCommit,
	stI metast.State,
) {
	// noop
	// if in the future we want to update the block's table with the apphash,
	// this is the best place to do it. We'll need to update the schema to grant
	// update permissions to the node user, though.
}

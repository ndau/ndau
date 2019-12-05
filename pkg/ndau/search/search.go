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
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/oneiro-ndev/ndau/pkg/query"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// SearchSysvarHistory returns value history for the given sysvar.
// The response is sorted by ascending block height, each entry is where the key's value changed.
//
// `minHeight` has inclusive semantics.
// Pass in 0,0 for the paging params to get the entire history.
func (client *Client) SearchSysvarHistory(
	sysvar string, minHeight uint64, limit int,
) (khr *query.SysvarHistoryResponse, err error) {
	khr = new(query.SysvarHistoryResponse)

	var rows pgx.Rows
	if limit > 0 {
		rows, err = client.Postgres.Query(
			context.Background(),
			"SELECT COALESCE(height, 0), value FROM systemvariables WHERE height>=$1 "+
				"ORDER BY height ASC NULLS FIRST LIMIT $2",
			minHeight, limit,
		)
	} else {
		rows, err = client.Postgres.Query(
			context.Background(),
			"SELECT COALESCE(height, 0), value FROM systemvariables WHERE height>=$1 "+
				"ORDER BY height ASC NULLS FIRST",
			minHeight,
		)
	}
	if err != nil {
		return nil, errors.Wrap(err, "querying system variables")
	}
	defer rows.Close()

	for rows.Next() {
		var height uint64
		var value []byte
		err = rows.Scan(&height, &value)
		if err != nil {
			err = errors.Wrap(err, "scanning sysvar row")
			return
		}
		khr.History = append(khr.History, query.SysvarHistoricalValue{
			Height: height,
			Value:  value,
		})
	}

	return
}

// SearchBlockHash returns the height of the given block hash.
// Returns 0 and no error if the given block hash was not found in the index.
func (client *Client) SearchBlockHash(blockHash string) (uint64, error) {
	var height uint64
	err := client.Postgres.QueryRow(
		context.Background(),
		"SELECT height FROM blocks WHERE hash=$1 LIMIT 1",
		blockHash,
	).Scan(&height)
	if err == pgx.ErrNoRows {
		err = nil
	}
	return height, err
}

// SearchTxHash returns tx data for the given tx hash.
func (client *Client) SearchTxHash(txHash string) (txvd TxValueData, err error) {
	err = client.Postgres.QueryRow(
		context.Background(),
		"SELECT height, sequence, fee, sib FROM transactions WHERE hash=$1 LIMIT 1",
		txHash,
	).Scan(&txvd.BlockHeight, &txvd.TxOffset, &txvd.Fee, &txvd.SIB)
	return
}

// SearchTxTypes returns tx data for a range of transactions on or before the given tx hash.
//
// This is hard to use properly, and is not particularly efficient. More specific
// searches should be preferred.
//
// If txHashOrHeight is "", this will return the latest page of transactions from the blockchain.
// txHashOrHeight can be a block height.  Transactions in and before that block are returned.
// If txTypes is empty, this will not filter on transaction name.
// If limit is non-positive, this will return results as if the page size is infinite.
func (client *Client) SearchTxTypes(
	txHashOrHeight string,
	txTypes []string,
	limit int,
) (lvd TxListValueData, err error) {
	// construct the query
	query := "SELECT height, sequence, fee, sib, hash FROM transactions "
	haswhere := false
	args := make([]interface{}, 0, 5)
	where := func(clause string) {
		if haswhere {
			query += "AND "
		} else {
			query += "WHERE "
			haswhere = true
		}
		query += clause
	}
	if txHashOrHeight != "" {
		where("(hash=$%d OR height::text=$%d) ")
		args = append(args, txHashOrHeight)
		query = fmt.Sprintf(query, len(args), len(args))
	}
	if len(txTypes) > 0 {
		where("(name = ANY ($%d)) ")
		args = append(args, txTypes)
		query = fmt.Sprintf(query, len(args))
	}
	query += "ORDER BY height DESC, sequence DESC "
	if limit > 0 {
		args = append(args, limit+1)
		query += fmt.Sprintf("LIMIT $%d ", len(args))
	}

	// perform the query
	var rows pgx.Rows
	rows, err = client.Postgres.Query(
		context.Background(),
		query,
		args...,
	)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	defer rows.Close()

	// build the results
	count := 0
	var hash string
	for rows.Next() {
		txvd := TxValueData{}
		err = rows.Scan(&txvd.BlockHeight, &txvd.TxOffset, &txvd.Fee, &txvd.SIB, &hash)
		if err != nil {
			err = errors.Wrap(err, "scanning transactions")
			return
		}
		if limit > 0 && count >= limit {
			lvd.NextTxHash = hash
			return
		}
		lvd.Txs = append(lvd.Txs, txvd)
		count++
	}
	return
}

// SearchAccountHistory returns an array of block height and txoffset pairs associated with the
// given account address.
//
// `afterHeight` has exclusive semantics
// Pass in 0, 0 for the paging params to get the entire history.
func (client *Client) SearchAccountHistory(
	addr string, afterHeight uint64, limit int,
) (ahr *AccountHistoryResponse, err error) {
	ahr = new(AccountHistoryResponse)

	var args = []interface{}{addr, afterHeight}
	query := "SELECT height, sequence, accounts.data->'balance' AS balance " +
		"FROM accounts JOIN transactions ON accounts.tx=transactions.id " +
		"WHERE accounts.address=$1 AND height>$2 "
	if limit > 0 {
		query += "LIMIT $3 "
		args = append(args, limit+1)
	}

	// perform the query
	var rows pgx.Rows
	rows, err = client.Postgres.Query(
		context.Background(),
		query,
		args...,
	)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	defer rows.Close()

	// build the results
	count := 0
	for rows.Next() {
		txvd := AccountTxValueData{}
		err = rows.Scan(&txvd.BlockHeight, &txvd.TxOffset, &txvd.Balance)
		if err != nil {
			err = errors.Wrap(err, "scanning transactions")
			return
		}
		if limit > 0 && count >= limit {
			ahr.More = true
			return
		}
		ahr.Txs = append(ahr.Txs, txvd)
		count++
	}
	return
}

// BlockTime returns the timestamp for the block at a given height
// returns the zero value and no error if the block is unknown
func (client *Client) BlockTime(height uint64) (ts math.Timestamp, err error) {
	var t time.Time
	err = client.Postgres.QueryRow(
		context.Background(),
		"SELECT block_time FROM blocks WHERE height=$1 LIMIT 1",
		height,
	).Scan(&t)
	if err != nil {
		err = errors.Wrap(err, "querying block time")
	}
	var zt time.Time
	if t == zt {
		return math.Timestamp(0), nil
	}
	ts, err = math.TimestampFrom(t)
	err = errors.Wrap(err, "converting to ndau time")
	return
}

// SearchMostRecentRegisterNode returns tx data for the most recent
// RegisterNode transactions for the given address.
//
// Returns the epoch date and no error if the node has never been registered.
func (client *Client) SearchMostRecentRegisterNode(address string) (ts math.Timestamp, err error) {
	var tts *time.Time
	err = client.Postgres.QueryRow(
		context.Background(),
		"SELECT block_time FROM "+
			"blocks INNER JOIN transactions "+
			"ON blocks.height=transactions.height "+
			"WHERE name='RegisterNode' AND (data->>'node')=$1 "+
			"ORDER BY transactions.height DESC, transactions.sequence DESC "+
			"LIMIT 1 ",
		address,
	).Scan(&tts)

	if err == pgx.ErrNoRows {
		err = nil
	}
	if err != nil {
		err = errors.Wrap(err, "querying db")
		return
	}

	if tts != nil {
		ts, err = math.TimestampFrom(*tts)
		err = errors.Wrap(err, "converting timestamp")
	}

	return
}

// SearchMarketPrice searches for market price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (client *Client) SearchMarketPrice(params PriceQueryParams) (PriceQueryResults, error) {
	return client.searchPrice(params, "marketprices")
}

// SearchTargetPrice searches for target price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (client *Client) SearchTargetPrice(params PriceQueryParams) (PriceQueryResults, error) {
	return client.searchPrice(params, "targetprices")
}

// searchPrice searches for market price records
//
// In the parameters:
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
func (client *Client) searchPrice(
	params PriceQueryParams,
	table string,
) (out PriceQueryResults, err error) {
	// this is a complicated query with a lot of parameterization, so we end up
	// doing quite a lot of string formatting, because postgres complains about
	// variables for the table name, and receiving a different number of parameters
	// than are called for in the query
	query := "SELECT price, b.height, block_time " +
		fmt.Sprintf("FROM %s AS p ", table) +
		"INNER JOIN transactions AS tx ON p.tx=tx.id " +
		"INNER JOIN blocks AS b on b.height=tx.height "
	haswhere := false
	args := make([]interface{}, 0, 5)
	where := func(clause string) {
		if haswhere {
			query += "AND "
		} else {
			query += "WHERE "
			haswhere = true
		}
		query += clause
	}
	if params.After.Timestamp > 0 {
		where("block_time>$%d ")
		args = append(args, params.After.Timestamp.AsTime())
		query = fmt.Sprintf(query, len(args))
	}
	if params.After.Height > 0 {
		where("b.height>$%d ")
		args = append(args, params.After.Height)
		query = fmt.Sprintf(query, len(args))
	}
	if params.Before.Timestamp > 0 {
		where("block_time<$%d ")
		args = append(args, params.Before.Timestamp.AsTime())
		query = fmt.Sprintf(query, len(args))
	}
	if params.Before.Height > 0 {
		where("b.height<$%d ")
		args = append(args, params.Before.Height)
		query = fmt.Sprintf(query, len(args))
	}
	query += "ORDER BY tx.height, tx.sequence "
	if params.Limit > 0 {
		query += "LIMIT $%d "
		args = append(args, params.Limit+1)
		query = fmt.Sprintf(query, len(args))
	}

	// perform the query
	var rows pgx.Rows
	rows, err = client.Postgres.Query(
		context.Background(),
		query,
		args...,
	)
	if err != nil {
		err = errors.Wrap(err, query)
		return
	}
	defer rows.Close()

	// build the results
	count := uint(0)
	for rows.Next() {
		item := PriceQueryResult{}
		var ts time.Time
		err = rows.Scan(&item.Price, &item.Height, &ts)
		if err != nil {
			err = errors.Wrap(err, "scanning prices")
			return
		}
		item.Timestamp, err = math.TimestampFrom(ts)
		if err != nil {
			err = errors.Wrap(err, "converting timestamp")
			return
		}
		if params.Limit > 0 && count >= params.Limit {
			out.More = true
			return
		}
		out.Items = append(out.Items, item)
		count++
	}
	return
}

// SearchDateRange returns the first and last block heights for the given ISO-3339 date range.
// The first is inclusive, the last is exclusive.
// Returns 0, 0, nil if no blocks lie within the specified range.
func (client *Client) SearchDateRange(f, l math.Timestamp) (first uint64, last uint64, err error) {
	err = client.Postgres.QueryRow(
		context.Background(),
		"SELECT "+
			"COALESCE(MIN(height), 0) AS first, "+
			"COALESCE(MAX(height), 0) AS last "+
			"FROM blocks WHERE block_time >= $1 AND block_time < $2",
		f.AsTime(), l.AsTime(),
	).Scan(&first, &last)
	return
}

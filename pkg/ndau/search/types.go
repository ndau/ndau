package search

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// Types common to indexing and searching.

import (
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

// HeightByBlockHashCommand is a QueryParams command for searching block height by block hash.
const HeightByBlockHashCommand = "heightbyblockhash"

// HeightByTxHashCommand is a QueryParams command for searching block height by tx hash.
const HeightByTxHashCommand = "heightbytxhash"

// HeightsByTxTypesCommand is a QueryParams command for searching block heights by tx types.
const HeightsByTxTypesCommand = "heightsbytxtypes"

// QueryParams is a json-friendly struct for passing query terms over endpoints.
type QueryParams struct {
	// App-specific command.
	Command string `json:"command"`

	// A block hash or tx hash (or any other kind of hash), depending on the command.
	Hash string `json:"hash"`

	// List of tx types, or any other format depending on the command.
	Types []string `json:"types"`

	// Useful for paging queries.
	Limit int `json:"limit"`
}

// SysvarHistoryParams is a json-friendly struct for the /sysvar/history endpoint.
type SysvarHistoryParams struct {
	Name        string `json:"name"`
	AfterHeight uint64 `json:"afterheight"`
	Limit       int    `json:"limit"`
}

// AccountHistoryParams is a json-friendly struct for the /account/history endpoint.
type AccountHistoryParams struct {
	Address     string `json:"addr"`
	AfterHeight uint64 `json:"afterheight"`
	Limit       int    `json:"limit"`
}

// AccountListParams is a json-friendly struct for the /account/list endpoint.
type AccountListParams struct {
	Address string `json:"addr"`
	After   string `json:"after"`
	Limit   int    `json:"limit"`
}

// RangeEndpoint is a json-friendly struct for choosing the end of a range
//
// At most one of (`Height`, `Timestamp`) should ever be set. If both are set,
// `Timestamp` takes precedence.
type RangeEndpoint struct {
	Height    uint64         `json:"block_height,omitempty"`
	Timestamp math.Timestamp `json:"timestamp,omitempty"`
}

// PriceQueryParams is a json-friendly struct for querying price history
//
// Before and After have exclusive semantics.
//
// The zero value of Before and After are treated as open-ended ranges.
// The zero value of Limit returns all results.
type PriceQueryParams struct {
	After  RangeEndpoint `json:"after,omitempty"`
	Before RangeEndpoint `json:"before,omitempty"`
	Limit  uint          `json:"limit,omitempty"`
}

// PriceQueryResult is a json-friendly struct returning price history data
type PriceQueryResult struct {
	Price     pricecurve.Nanocent `json:"price_nanocents"`
	PriceS    string              `json:"price,omitempty"`
	Height    uint64              `json:"block_height"`
	Timestamp math.Timestamp      `json:"timestamp"`
}

// PriceQueryResults encapsulates a set of price history data
//
// It is _not_ json-friendly; More should be replaced with Next at the API level
// More is true when more results exist than were returned
type PriceQueryResults struct {
	Items []PriceQueryResult `json:"-"`
	More  bool               `json:"-"`
}

// TxValueData is used for data about a particular transaction.
type TxValueData struct {
	BlockHeight uint64 `json:"height" msg:"h"`
	TxOffset    int    `json:"offset" msg:"o"`
	Fee         uint64 `json:"fee" msg:"f"`
	SIB         uint64 `json:"sib" msg:"s"`
}

// TxListValueData is used for data about a list of transactions.
type TxListValueData struct {
	Txs        []TxValueData `json:"txs" msg:"t"`
	NextTxHash string        `json:"next" msg:"n"`
}

// AccountTxValueData is like TxValueData that stores account balance at the associated block.
// We could index a Ref target hash, but that would use more space than just storing the balance.
type AccountTxValueData struct {
	BlockHeight uint64     `msg:"h"`
	TxOffset    int        `msg:"o"`
	Balance     types.Ndau `msg:"b"`
}

// AccountHistoryResponse is the return value from the account history endpoint.
type AccountHistoryResponse struct {
	Txs  []AccountTxValueData `msg:"t"`
	More bool                 `msg:"m"`
}

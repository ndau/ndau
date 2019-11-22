package query

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp -io=0

// SysvarHistoricalValue is a value, and the height at which it was set
type SysvarHistoricalValue struct {
	Height uint64 `json:"height"`
	Value  []byte `json:"value"`
}

// SysvarHistoryResponse returns the history of a key over time.
//
// For compactness, history is compressed, and records are only returned for those
// blocks on which the value changed.
type SysvarHistoryResponse struct {
	History []SysvarHistoricalValue `json:"history"`
}

// Summary is the return value from the /summary endpoint
type Summary struct {
	BlockHeight      uint64
	TotalNdau        types.Ndau
	NumAccounts      int
	TotalRFE         types.Ndau
	TotalIssue       types.Ndau
	TotalBurned      types.Ndau
	TotalCirculation types.Ndau
}

// SidechainTxExistsQuery specifies a particular sidechain tx
type SidechainTxExistsQuery struct {
	SidechainID byte
	Source      address.Address
	TxHash      string
}

// AccountListQueryResponse is the return value from the /accountlist endpoint
type AccountListQueryResponse struct {
	NumAccounts int
	FirstIndex  int
	After       string
	NextAfter   string
	Accounts    []string
}

// DelegateList lists the accounts delegated to a particular node
type DelegateList struct {
	Node      address.Address
	Delegated []address.Address
}

// DelegatesResponse is the return value from the /delegates endpoint
//
// Note that this is _not_ a standard MSGP-able struct; it must instead
// be marshalled and unmarshalled using msgp.(Un)MarshalIntf methods
type DelegatesResponse []DelegateList

// SIBResponse is the return value from the /sib endpoint
//
// This includes market and target price values so end users can check our math
type SIBResponse struct {
	SIB         eai.Rate
	TargetPrice pricecurve.Nanocent
	MarketPrice pricecurve.Nanocent
	FloorPrice  pricecurve.Nanocent
}

// SysvarsRequest is the request value for the /sysvars endpoint
//
// If set, only the named sysvars are returned
type SysvarsRequest []string

// SysvarsResponse is the return value from the /sysvars endpoint
type SysvarsResponse map[string][]byte

// NodeExtra has managed data which would not otherwise be captured by the
// JSON format of a node
type NodeExtra struct {
	Node         backing.Node    `json:"node"`
	Registration types.Timestamp `json:"registration"`
}

// NodesResponse is the return value from the /nodes endpoint
type NodesResponse map[string]NodeExtra

// DateRangeRequest is used for passing date range query terms over endpoints.
type DateRangeRequest struct {
	FirstTimestamp types.Timestamp
	LastTimestamp  types.Timestamp
}

// DateRangeResult is used for returning search results for the date range endpoint.
type DateRangeResult struct {
	FirstHeight uint64
	LastHeight  uint64
}

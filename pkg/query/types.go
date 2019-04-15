package query

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

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
}

// SysvarsRequest is the request value for the /sysvars endpoint
//
// If set, only the named sysvars are returned
type SysvarsRequest []string

// SysvarsResponse is the return value from the /sysvars endpoint
type SysvarsResponse map[string][]byte

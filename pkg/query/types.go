package query

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

// Summary is the return value from the /summary endpoint
type Summary struct {
	BlockHeight      uint64
	TotalNdau        types.Ndau
	NumAccounts      int
	TotalRFE         types.Ndau
	TotalIssue       types.Ndau
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
	PageSize    int
	PageIndex   int
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

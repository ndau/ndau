package query

import "github.com/oneiro-ndev/ndaumath/pkg/types"

//go:generate msgp

// Summary is the return value from the /summary endpoint
type Summary struct {
	BlockHeight uint64
	TotalNdau   types.Ndau
	NumAccounts int
}

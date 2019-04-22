package backing

import math "github.com/oneiro-ndev/ndaumath/pkg/types"

// generate msgp interface implementations for AccountData and supporting structs
// we can't generate the streaming interfaces, unfortunately, because the
// signature.* types don't implement those
//go:generate msgp -io=0

// generate noms marshaler implementations for appropriate types
//go:generate go run $GOPATH/src/github.com/oneiro-ndev/generator/cmd/nomsify $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/backing
//go:generate find $GOPATH/src/github.com/oneiro-ndev/ndau/pkg/ndau/backing -name "*noms_gen*.go" -maxdepth 1 -exec goimports -w {} ;
//nomsify Hold

// Hold tracks a portion of this account's ndau which cannot currently be spent
type Hold struct {
	Qty    math.Ndau       `json:"qty" chain:"81,Hold_Quantity"`
	Expiry *math.Timestamp `json:"expiry" chain:"82,Hold_Expiry"`
	Txhash *string         `json:"tx_hash" chain:"83,Hold_TxHash"`
	Stake  *StakeData      `json:"stake" chain:"84,Hold_Stake"`
}

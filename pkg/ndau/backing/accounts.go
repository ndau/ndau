package backing

import (
	"github.com/attic-labs/noms/go/marshal"
	nt "github.com/attic-labs/noms/go/types"

	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	address "github.com/oneiro-ndev/ndaunode/pkg/node.address"
	util "github.com/oneiro-ndev/noms-util"
)

// An Account is a map from an Address to an AccountData struct
type Account nt.Map

// Lock keeps track of an account's Lock information
type Lock struct {
	Duration math.Duration
	// if a lock has not been notified, this is nil
	NotifiedOn *math.Timestamp `noms:",omitempty"`
}

var _ marshal.Marshaler = (*Lock)(nil)
var _ marshal.Unmarshaler = (*Lock)(nil)

// MarshalNoms implements Marshaler for lock
func (l Lock) MarshalNoms(vrw nt.ValueReadWriter) (val nt.Value, err error) {
	return marshal.Marshal(vrw, l.toNomsLock())
}

// UnmarshalNoms implements Unmarshaler for lock
func (l *Lock) UnmarshalNoms(v nt.Value) error {
	nl := nomsLock{}
	err := marshal.Unmarshal(v, &nl)
	if err != nil {
		return err
	}
	l.fromNomsLock(nl)
	return nil
}

type nomsLock struct {
	Duration   util.Int
	IsNotified bool
	NotifiedOn util.Int
}

func (l Lock) toNomsLock() nomsLock {
	nl := nomsLock{
		Duration:   util.Int(l.Duration),
		IsNotified: l.NotifiedOn != nil,
	}
	if l.NotifiedOn != nil {
		nl.NotifiedOn = util.Int(*l.NotifiedOn)
	}
	return nl
}

func (l *Lock) fromNomsLock(nl nomsLock) {
	l.Duration = math.Duration(nl.Duration)
	if nl.IsNotified {
		n := math.Timestamp(nl.NotifiedOn)
		l.NotifiedOn = &n
	} else {
		l.NotifiedOn = nil
	}
}

// Stake keeps track of an account's staking information
type Stake struct {
	Point   math.Timestamp
	Address address.Address
}

// Escrow tracks a single transaction of incoming escrow
type Escrow struct {
	Qty math.Ndau
	// Expiry is when these funds are available to be sent
	Expiry math.Timestamp
}

// EscrowSettings tracks the escrow settings for outbound transactions
type EscrowSettings struct {
	Duration  math.Duration
	ChangesAt *math.Timestamp
	Next      *math.Duration
}

// AccountData contains all the information the node needs to take action on a particular account.
//
// See the whitepaper: https://github.com/oneiro-ndev/whitepapers/blob/master/node_incentives/transactions.md#wallet-data
type AccountData struct {
	Balance        math.Ndau
	Lock           *Lock
	Stake          *Stake
	UpdatePoint    math.Timestamp
	Sequence       uint64
	Escrows        []Escrow
	EscrowSettings EscrowSettings
}

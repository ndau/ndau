package backing

import (
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
)

//go:generate msgp

// generate noms marshaler implementations for appropriate types
//nomsify StakeData StakeRules

// StakeData keeps track of a particular stake or costake
type StakeData struct {
	Point     math.Timestamp  `json:"point" chain:"101,Stake_Point"`
	RulesAcct address.Address `json:"rules_acct" chain:"102,Stake_RulesAcct"`
	StakeTo   address.Address `json:"stake_to" chain:"103,Stake_To"`
}

// StakeRules keeps track of data associated with Rules accounts
type StakeRules struct {
	Script  []byte            `json:"script" chain:"120,StakeRules_Script"`
	Inbound map[string]uint64 `json:"inbound" chain:"121,StakeRules_Inbound"`
}

// HoldSum returns the sum of held ndau in this account
func (a AccountData) HoldSum() math.Ndau {
	m := math.Ndau(0)
	for _, d := range a.Holds {
		m += d.Qty
	}
	return m
}

// VisitStakesOutbound visits each stake in this account
//
// If the visitor ever returns true, iteration stops.
func (a AccountData) VisitStakesOutbound(visit func(Hold) (stop bool)) {
	for _, d := range a.Holds {
		if d.Stake != nil {
			stop := visit(d)
			if stop {
				break
			}
		}
	}
}

// StakeOutboundSum returns the sum of ndau staked from this account
func (a AccountData) StakeOutboundSum() math.Ndau {
	m := math.Ndau(0)
	a.VisitStakesOutbound(func(h Hold) bool {
		m += h.Qty
		return false
	})
	return m
}

// PrimaryStake returns the stake data for a primary stake to a rules account
//
// If it doesn't exist, returns nil.
func (a AccountData) PrimaryStake(rules address.Address) (ps *Hold) {
	s := rules.String()
	a.VisitStakesOutbound(func(h Hold) bool {
		if h.Stake.StakeTo.String() == s && h.Stake.RulesAcct.String() == s {
			ps = &h
			return true
		}
		return false
	})
	return
}

// AggregateStake returns the aggregate stake data from a target account to a rules account
//
// The only way in which this differs from target.PrimaryStake is that the Qty is suitably
// adjusted to include the contributions of all costakers.
//
// If target isn't a primary staker to the rules account, returns nil.
func (s State) AggregateStake(target, rules address.Address) (ps *Hold) {
	a := s.Accounts[target.String()]
	ps = a.PrimaryStake(rules)
	if ps != nil {
		for costaker := range a.Costakers[rules.String()] {
			c := s.Accounts[costaker]
			c.VisitStakesOutbound(func(h Hold) (stop bool) {
				if h.Stake.StakeTo == target && h.Stake.RulesAcct == rules {
					ps.Qty += h.Qty
				}
				return false
			})
		}
	}
	return
}

// IsSelfStakedTo is true when the target is self-staked to the rules account
func (a AccountData) IsSelfStakedTo(rules address.Address) bool {
	return a.PrimaryStake(rules) != nil
}
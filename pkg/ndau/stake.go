package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"fmt"

	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signed"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// NodeStakers returns all stakers and costakers of a node and their total stake
func (app *App) NodeStakers(node address.Address) (map[string]math.Ndau, error) {
	var nra address.Address
	err := app.System(sv.NodeRulesAccountAddressName, &nra)
	if err != nil {
		return nil, errors.Wrap(err, "getting node rules account address")
	}

	stakers := make(map[string]math.Ndau)

	// first, add in the primary stake and self-stakes
	updateStakersFor := func(addr address.Address) {
		acct, _ := app.getAccount(addr)
		acct.VisitStakesOutbound(func(h backing.Hold) bool {
			hs := h.Stake
			if hs.RulesAcct == nra && (hs.StakeTo == node || (addr == node && hs.StakeTo == nra)) {
				stakers[addr.String()] += h.Qty
			}
			return false
		})
	}
	updateStakersFor(node)

	nodeAcct, _ := app.getAccount(node)
	for costaker := range nodeAcct.Costakers[nra.String()] {
		caddr, err := address.Validate(costaker)
		if err != nil {
			continue
			// this should never happen but if it does, just continue
		}
		print(caddr.String())
		// Don't double-count stakes (primary or self) to ourself
		if caddr != node {
			print("Not a duplicate - add to total")
			updateStakersFor(caddr)
		}
	}

	return stakers, nil
}

// Stake updates the state to handle staking an account to another
//
// It is assumed that all necessary validation has already been performed. In
// particular, this function does not attempt to construct or run the chaincode
// context for the rules account.
//
// This function returns a function suitable for calling within app.UpdateState
func (app *App) Stake(
	qty math.Ndau,
	target, stakeTo, rules address.Address,
	tx metatx.Transactable,
) func(metast.State) (metast.State, error) {
	return func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		targetAcct, _ := app.getAccount(target)

		targetAcct.UpdateRecourses(app.BlockTime())

		ab, err := targetAcct.AvailableBalance()
		if err != nil {
			return st, errors.Wrap(err, "target available balance")
		}

		if ab < qty {
			return st, fmt.Errorf("stake: insufficient target available balance: have %d, need %d", ab, qty)
		}

		rulesAcct, _ := app.getAccount(rules)
		if rulesAcct.StakeRules == nil {
			return st, fmt.Errorf("stake: rules must be a rules account")
		}

		isPrimary := stakeTo == rules

		if isPrimary {
			ps := targetAcct.PrimaryStake(rules)
			if ps != nil {
				return st, fmt.Errorf("stake: target cannot have more than 1 primary stake to a rules account")
			}
		}

		// update 3 places where we keep track of rules info:
		// - outbound stake list
		// - rules inbounds (if primary)
		// - costakers list (if applicable)
		hold := backing.Hold{
			Qty: qty,
			Stake: &backing.StakeData{
				Point:     app.BlockTime(),
				RulesAcct: rules,
				StakeTo:   stakeTo,
			},
		}
		if tx != nil {
			hold.Txhash = metatx.Hash(tx)
		}
		targetAcct.Holds = append(targetAcct.Holds, hold)
		st.Accounts[target.String()] = targetAcct

		if isPrimary {
			rulesAcct, _ := app.getAccount(rules)
			rulesAcct.StakeRules.Inbound[target.String()]++
			st.Accounts[rules.String()] = rulesAcct
		} else {
			stakeToAcct, _ := app.getAccount(stakeTo)
			rulesCostakers := stakeToAcct.Costakers[rules.String()]
			if rulesCostakers == nil {
				rulesCostakers = make(map[string]uint64)
			}
			rulesCostakers[target.String()]++
			if stakeToAcct.Costakers == nil {
				stakeToAcct.Costakers = make(map[string]map[string]uint64)
			}
			stakeToAcct.Costakers[rules.String()] = rulesCostakers
			st.Accounts[stakeTo.String()] = stakeToAcct
		}

		return st, nil
	}
}

// Unstake updates the state to handle unstaking an account
//
// It is assumed that all necessary validation has already been performed. In
// particular, this function does not attempt to construct or run the chaincode
// context for the rules account.
//
// This function returns a function suitable for calling within app.UpdateState
func (app *App) Unstake(
	qty math.Ndau,
	target, stakeTo, rules address.Address,
	retainFor math.Duration,
) func(metast.State) (metast.State, error) {
	return app.UnstakeAndBurn(qty, 0, target, stakeTo, rules, retainFor, false)
}

// UnstakeAndBurn updates the state to handle unstaking an account with some
// amount burned. Burn is expressed as a fraction.
//
// It is assumed that all necessary validation has already been performed. In
// particular, this function does not attempt to construct or run the chaincode
// context for the rules account.
//
// This function returns a function suitable for calling within app.UpdateState
func (app *App) UnstakeAndBurn(
	qty math.Ndau, burn uint8,
	target, stakeTo, rules address.Address,
	retainFor math.Duration,
	recursive bool,
) func(metast.State) (metast.State, error) {
	return func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		targetAcct, _ := app.getAccount(target)

		// update 3 places where we keep track of rules info:
		// - outbound stake list
		// - rules inbounds (if primary)
		// - costakers list (if applicable)
		found := 0
		for idx, hold := range targetAcct.Holds {
			if (recursive || hold.Qty == qty) &&
				hold.Stake != nil &&
				hold.Stake.RulesAcct == rules &&
				// If the staker we're looking at right now is the
				// target, then StakeTo must be the Rules account to match.
				// Otherwise, it must be what we're staking to.
				(target == stakeTo &&
					hold.Stake.StakeTo == rules ||
					hold.Stake.StakeTo == stakeTo) {

				found++

				burned, err := signed.MulDiv(int64(hold.Qty), int64(burn), resolveStakeDenominator)
				if err != nil {
					return nil, errors.Wrap(err, "computing burned qty for "+target.String())
				}
				nburned := math.Ndau(burned)
				targetAcct.Balance -= nburned
				st.TotalBurned += nburned

				if retainFor == 0 {
					// quickly remove this element by replacing it with the final one
					targetAcct.Holds[idx] = targetAcct.Holds[len(targetAcct.Holds)-1]
					targetAcct.Holds = targetAcct.Holds[:len(targetAcct.Holds)-1]
				} else {
					// convert this stake hold into a normal hold with an appropriate
					// retention period
					hold.Stake = nil
					uo := app.BlockTime().Add(retainFor)
					hold.Expiry = &uo
					// reduce the qty by the amount burned
					hold.Qty -= nburned
					targetAcct.Holds[idx] = hold
				}

				if !recursive {
					break
				}
			}
		}
		if found == 0 {
			// didn't discover a matching hold
			return nil, errors.New("No matching hold found in " + target.String())
		}
		st.Accounts[target.String()] = targetAcct

		rulesAcct, _ := app.getAccount(rules)
		stakeToAcct, _ := app.getAccount(stakeTo)
		rulesCostakers := stakeToAcct.Costakers[rules.String()]
		if stakeTo == rules || target == stakeTo {
			rulesAcct.StakeRules.Inbound[target.String()]--
			if rulesAcct.StakeRules.Inbound[target.String()] == 0 {
				delete(rulesAcct.StakeRules.Inbound, target.String())
			}
			if recursive {
				for costakerS := range rulesCostakers {
					costaker, err := address.Validate(costakerS)
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("invalid costaker %s for primary %s", costakerS, target))
					}
					stI, err = app.UnstakeAndBurn(qty, burn, costaker, stakeTo, rules, retainFor, true)(st)
					if err != nil {
						return nil, errors.Wrap(err, fmt.Sprintf("unstaking costaker %s for primary %s", costakerS, target))
					}
				}
			}
		} else {
			if rulesCostakers != nil {
				rulesCostakers[target.String()]--
				if rulesCostakers[target.String()] == 0 {
					delete(rulesCostakers, target.String())
				}
				if len(rulesCostakers) == 0 {
					delete(stakeToAcct.Costakers, rules.String())
				} else {
					stakeToAcct.Costakers[rules.String()] = rulesCostakers
				}
			}
		}

		st.Accounts[stakeTo.String()] = stakeToAcct
		st.Accounts[rules.String()] = rulesAcct

		// JSG the above might have modified total ndau in circulation, so recalculate SIB
		if app.IsFeatureActive("AllRFEInCirculation") {
			sib, target, err := app.calculateCurrentSIB(st, -1, -1)
			if err != nil {
				return st, err
			}
			st.SIB = sib
			st.TargetPrice = target
		}

		return st, nil
	}
}

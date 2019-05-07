package ndau

import (
	"fmt"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
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
		updateStakersFor(caddr)
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
		stakeToAcct, _ := app.getAccount(stakeTo)
		rulesAcct, _ := app.getAccount(rules)

		targetAcct.UpdateSettlements(app.BlockTime())

		ab, err := targetAcct.AvailableBalance()
		if err != nil {
			return st, errors.Wrap(err, "target available balance")
		}

		if ab < qty {
			return st, fmt.Errorf("stake: insufficient target available balance: have %d, need %d", ab, qty)
		}

		if rulesAcct.StakeRules == nil {
			return st, fmt.Errorf("stake: rules must be a rules account")
		}

		isPrimary := stakeTo == rules

		if isPrimary {
			ps := targetAcct.PrimaryStake(rules)
			if ps != nil {
				return st, fmt.Errorf("stake: cannot have more than 1 primary stake to a rules account")
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

		if isPrimary {
			rulesAcct.StakeRules.Inbound[target.String()]++
		} else {
			rulesCostakers := stakeToAcct.Costakers[rules.String()]
			if rulesCostakers == nil {
				rulesCostakers = make(map[string]uint64)
			}
			rulesCostakers[target.String()]++
			if stakeToAcct.Costakers == nil {
				stakeToAcct.Costakers = make(map[string]map[string]uint64)
			}
			stakeToAcct.Costakers[rules.String()] = rulesCostakers
		}

		st.Accounts[target.String()] = targetAcct
		st.Accounts[stakeTo.String()] = stakeToAcct
		st.Accounts[rules.String()] = rulesAcct

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
) func(metast.State) (metast.State, error) {
	return func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		targetAcct, _ := app.getAccount(target)
		stakeToAcct, _ := app.getAccount(stakeTo)
		rulesAcct, _ := app.getAccount(rules)

		// update 3 places where we keep track of rules info:
		// - outbound stake list
		// - rules inbounds (if primary)
		// - costakers list (if applicable)
		for idx, hold := range targetAcct.Holds {
			if hold.Qty == qty && hold.Stake != nil && hold.Stake.StakeTo == stakeTo && hold.Stake.RulesAcct == rules {
				// quickly remove this element by replacing it with the final one
				targetAcct.Holds[idx] = targetAcct.Holds[len(targetAcct.Holds)-1]
				targetAcct.Holds = targetAcct.Holds[:len(targetAcct.Holds)-1]
				break
			}
		}

		if stakeTo == rules {
			rulesAcct.StakeRules.Inbound[target.String()]--
			if rulesAcct.StakeRules.Inbound[target.String()] == 0 {
				delete(rulesAcct.StakeRules.Inbound, target.String())
			}
		} else {
			rulesCostakers := stakeToAcct.Costakers[rules.String()]
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

		st.Accounts[target.String()] = targetAcct
		st.Accounts[stakeTo.String()] = stakeToAcct
		st.Accounts[rules.String()] = rulesAcct

		return st, nil
	}
}

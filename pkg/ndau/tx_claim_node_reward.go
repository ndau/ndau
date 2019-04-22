package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ClaimNodeReward) GetAccountAddresses() []string {
	return []string{tx.Node.String()}
}

// Validate implements metatx.Transactable
func (tx *ClaimNodeReward) Validate(appI interface{}) error {
	app := appI.(*App)

	_, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	timeout := math.Duration(0)
	err = app.System(sv.NodeRewardNominationTimeoutName, &timeout)
	if err != nil {
		return err
	}

	state := app.GetState().(*backing.State)

	if app.BlockTime().Compare(state.LastNodeRewardNomination.Add(timeout)) > 0 {
		return fmt.Errorf(
			"too late: NominateNodeReward @ %s expired after %s; currently %s",
			state.LastNodeRewardNomination,
			timeout,
			app.BlockTime(),
		)
	}

	if state.NodeRewardWinner == nil {
		return errors.New("no node reward winner; nobody has a valid claim")
	}

	if *state.NodeRewardWinner != tx.Node {
		return fmt.Errorf("winner was %s not %s", state.NodeRewardWinner, tx.Node)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ClaimNodeReward) Apply(appI interface{}) error {
	app := appI.(*App)
	err := app.applyTxDetails(tx)
	if err != nil {
		return err
	}

	costakers, err := app.NodeStakers(tx.Node)
	if err != nil {
		return errors.Wrap(err, "ClaimNodeReward")
	}

	return app.UpdateState(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		// We never want this tx to fail because we want all node payouts to be paid to someone.
		// If there are errors, they will be resolved in favor of the node operator.
		// That sounds unfair but the users will swiftly realize that their payouts aren't
		// coming properly and they'll punish the node op by moving their stake.
		// We need a place to hold any errors we may generate.
		// Because we want to dedup them, we just store them as strings and dump them at the end.
		allErrs := make(map[string]struct{})
		// this just makes the lines a little prettier
		xx := struct{}{}

		// work on distributing the awards list, and accumulate any errors we see. We want to
		// be sure that we always distribute all of the node reward, so we're never going to exit
		// this function early, and all errors are logged but not returned.

		rv := make(vm.List, 0)
		avm, err := BuildVMForNodeDistribution(
			state.Nodes[tx.Node.String()].DistributionScript,
			tx.Node,
			costakers,
			state.Accounts,
			state.UnclaimedNodeReward,
			app.BlockTime(),
		)
		// ugly nested ifs here but it's what we need to do
		// basically, if we have an egregious error anywhere up here, we'll skip trying
		// to evaluate the results and just give everything to the node op.
		if err != nil {
			allErrs[errors.Wrap(err, "constructing chaincode vm").Error()] = xx
		} else {
			err = avm.Run(nil)
			if err != nil {
				stackTop, err2 := avm.Stack().Peek()
				if err2 != nil {
					allErrs[fmt.Sprintf(
						"while logging an error, another occurred\noriginal: %s\nsecondary: %s",
						err, err2,
					)] = xx
				} else {
					allErrs[errors.Wrap(err, fmt.Sprintf("running chaincode vm (stack top: %s)", stackTop)).Error()] = xx
				}
			} else {
				// expected top of stack on successful chaincode completion:
				// list of structs, whose field 10 is the amount returned
				// note: this list may be sorted in arbitrary order, and sort order
				// may differ from our input list.
				rv, err = avm.Stack().PopAsListOfStructs(10)
				if err != nil {
					allErrs[errors.Wrap(err, "pop chaincode return value as list of structs").Error()] = xx
				}
			}
		}

		// Now, if the chaincode actually ran and returned something useful we can iterate through the list;
		// otherwise nothing will happen here and we'll just fall through to giving everything to the node
		// operator (which will eventually show up in the statistics).
		for _, v := range rv {
			s := v.(*vm.Struct)   // no err handling because PALOS promised
			value, _ := s.Get(10) // PALOS already checked that the field exists
			nV, isNum := value.(vm.Number)
			if !isNum {
				allErrs["chaincode returned non-numeric value for costaker disbursal"] = xx
				continue
			}
			n := nV.AsInt64()
			if n < 0 {
				allErrs["chaincode attempted to return negative reward for costaker"] = xx
				continue
			}
			addrV, err := s.Get(1) // addresses are in field 1
			if err != nil {
				allErrs[errors.Wrap(err, "getting address from chaincode return value").Error()] = xx
				continue
			}
			addrB, isBytes := addrV.(*vm.Bytes)
			if !isBytes {
				allErrs[errors.Wrap(err, "chaincode changed address type").Error()] = xx
				continue
			}
			addr := addrB.String()
			addrA, err := address.Validate(addr)
			if err != nil {
				allErrs[errors.Wrap(err, "chaincode messed up costaker address").Error()] = xx
				continue
			}

			if _, ok := costakers[addr]; !ok {
				allErrs["chaincode attempted to reward a non-costaker"] = xx
				continue
			}

			// awards are disbursed first-come, first-serve to the list of
			// costakers returned by the disbursement script.
			// If the script attempts to disburse more than the actual
			// total reward, rewards are truncated and subsequent costakers
			// are just out of luck.
			award := math.Ndau(0)
			final := false
			if math.Ndau(n) <= state.UnclaimedNodeReward {
				award = math.Ndau(n)
			} else {
				award = state.UnclaimedNodeReward
				final = true
			}
			state.UnclaimedNodeReward -= award
			_, err = state.PayReward(addrA, award, app.BlockTime(), app.getDefaultSettlementDuration(), false)
			if err != nil {
				allErrs[err.Error()] = xx
			}
			if final {
				break
			}
		}

		// if after disbursement to costakers there remains some node reward,
		// it goes to the node
		if state.UnclaimedNodeReward > 0 {
			_, err = state.PayReward(tx.Node, state.UnclaimedNodeReward, app.BlockTime(), app.getDefaultSettlementDuration(), false)
			if err != nil {
				allErrs[err.Error()] = xx
			}
			state.UnclaimedNodeReward = math.Ndau(0)
		}

		// we now have a (deduped) collection of errors that occurred, but we can't return them
		// since we always want to be sure that our payouts are applied, even if it's only
		// to the node operator. So instead we just log them.
		if len(allErrs) > 0 {
			// we need a logger
			logger := app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"tx":        "ClaimNodeReward",
				"node":      tx.Node.String(),
				"script":    state.Nodes[tx.Node.String()].DistributionScript,
				"blockTime": app.BlockTime(),
			})
			for e := range allErrs {
				logger.WithField("text", e).Error("error during node reward payout")
			}

		}

		return state, nil
	})
}

// GetSource implements sourcer
func (tx *ClaimNodeReward) GetSource(*App) (address.Address, error) {
	return tx.Node, nil
}

// GetSequence implements sequencer
func (tx *ClaimNodeReward) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements signeder
func (tx *ClaimNodeReward) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ClaimNodeReward) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

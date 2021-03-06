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
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

const resolveStakeDenominator = 255

// Validate implements metatx.Transactable
func (tx *ResolveStake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.Rules.String())
	if err != nil {
		return errors.Wrap(err, "rules")
	}

	// business rule: this field is interpreted as fraction
	// with implied denominator resolveStakeDenominator, indicating the disposition of
	// that portion of the staked amount.
	//
	// Note to future maintainers: when the value is less than 1, the remainder
	// of the stake is returned immediately to the stakers as spendable.
	// There is no cooldown because, once a human gets involved to issue
	// this tx, the stake is resolved and there is nothing further to do about it.
	if uint(tx.Burn) > resolveStakeDenominator {
		return fmt.Errorf("Burn must be <= %d", resolveStakeDenominator)
	}

	app := appI.(*App)

	rules, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if rules.StakeRules == nil {
		return fmt.Errorf("%s is not a rules account", tx.Rules)
	}

	target, _ := app.getAccount(tx.Target)

	if h := target.PrimaryStake(tx.Rules); h == nil {
		return fmt.Errorf("%s is not a primary staker to %s", tx.Target, tx.Rules)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ResolveStake) Apply(appI interface{}) error {
	app := appI.(*App)

	// all state changes get applied or none do
	return app.UpdateState(app.applyTxDetails(tx), func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		npayment, err := tx.calculatePayment(st)
		if err != nil {
			return st, err
		}

		target := st.Accounts[tx.Target.String()]
		if npayment > target.Balance {
			return st, fmt.Errorf(
				"invalid rules validation chaincode: payment (%d) exceeds balance (%d) for %s",
				npayment, target.Balance, tx.Target,
			)
		}

		rules := st.Accounts[tx.Rules.String()]
		rules.Balance += npayment
		st.Accounts[tx.Rules.String()] = rules

		// payment must come out of staked holds before U&B
		holdsFound := 0
		for idx, hold := range target.Holds {
			if hold.Stake != nil &&
				hold.Stake.RulesAcct == tx.Rules &&
				(hold.Stake.StakeTo == tx.Rules || hold.Stake.StakeTo == tx.Target) {
				holdsFound++
				if hold.Qty >= npayment {
					hold.Qty -= npayment
					target.Balance -= npayment
					npayment = 0
					break
				} else {
					hold.Qty = 0
					target.Balance -= hold.Qty
					npayment -= hold.Qty
				}
				target.Holds[idx] = hold
			}
		}
		if npayment > 0 {
			return st, fmt.Errorf(
				"in %d holds found, insufficient held ndau for payment from %s; %d remaining",
				holdsFound, tx.Target, npayment,
			)
		}

		st.Accounts[tx.Target.String()] = target

		return app.UnstakeAndBurn(0, tx.Burn, tx.Target, tx.Target, tx.Rules, 0, true)(st)
	})
}

// GetSource implements Sourcer
func (tx *ResolveStake) GetSource(*App) (address.Address, error) {
	return tx.Rules, nil
}

// GetSequence implements Sequencer
func (tx *ResolveStake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ResolveStake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ResolveStake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

func (tx *ResolveStake) calculatePayment(st *backing.State) (math.Ndau, error) {
	vm, err := BuildVMForRulesValidation(tx, st)
	if err != nil {
		return 0, errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return 0, errors.Wrap(err, "running rules validation vm")
	}
	payment, err := vm.Stack().PopAsInt64()
	if err != nil {
		return 0, errors.Wrap(err, "getting return code from rules validation vm")
	}
	return math.Ndau(payment), nil
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ResolveStake) GetAccountAddresses(app *App) ([]string, error) {
	return []string{tx.Target.String(), tx.Rules.String()}, nil
}

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

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Stake) Validate(appI interface{}) error {
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err, "target")
	}
	_, err = address.Validate(tx.StakeTo.String())
	if err != nil {
		return errors.Wrap(err, "stake_to")
	}
	_, err = address.Validate(tx.Rules.String())
	if err != nil {
		return errors.Wrap(err, "rules")
	}

	app := appI.(*App)
	target, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("target does not exist")
	}

	if tx.StakeTo == tx.Rules {
		ps := target.PrimaryStake(tx.Rules)
		if ps != nil {
			return fmt.Errorf("stake: cannot have more than 1 primary stake to a rules account")
		}
	}

	txFee, err := app.calculateTxFee(tx)
	if err != nil {
		return errors.Wrap(err, "calculating tx fee")
	}

	requiredBalance, err := tx.Qty.Add(txFee)
	if err != nil {
		return errors.Wrap(err, "calculating required balance")
	}

	if target.Balance.Compare(requiredBalance) < 0 {
		return fmt.Errorf("target has insufficient balance: have %s ndau, need %s", target.Balance, requiredBalance)
	}

	_, hasNode := app.getAccount(tx.StakeTo)
	if !hasNode {
		return errors.New("Node does not exist")
	}

	vm, err := BuildVMForRulesValidation(tx, app.GetState().(*backing.State))
	if err != nil {
		return errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return errors.Wrap(err, "running rules validation vm")
	}
	returncode, err := vm.Stack().PopAsInt64()
	if err != nil {
		return errors.Wrap(err, "getting return code from rules validation vm")
	}
	if returncode != 0 {
		return fmt.Errorf("rules validation script returned code %d", returncode)
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Stake) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(
		app.applyTxDetails(tx),
		app.Stake(tx.Qty, tx.Target, tx.StakeTo, tx.Rules, tx))
}

// GetSource implements Sourcer
func (tx *Stake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Stake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Stake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Stake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Stake) GetAccountAddresses(app *App) ([]string, error) {
	return []string{tx.Target.String(), tx.StakeTo.String(), tx.Rules.String()}, nil
}

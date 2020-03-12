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

	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Unstake) Validate(appI interface{}) error {
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
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return errors.Wrap(err, "sequence")
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

	if app.GetState().(*backing.State).IsActiveNode(tx.Target) {
		return errors.New("may not unstake an active node")
	}

	return err
}

// Apply implements metatx.Transactable
func (tx *Unstake) Apply(appI interface{}) error {
	app := appI.(*App)

	// recalculate the validation rules. This time we're not interested in
	// the stack top, but its second value. If a second value is present,
	// it's a duration to retain the hold for from the block time.
	vm, err := BuildVMForRulesValidation(tx, app.GetState().(*backing.State))
	if err != nil {
		return errors.Wrap(err, "building rules validation vm")
	}
	err = vm.Run(nil)
	if err != nil {
		return errors.Wrap(err, "running rules validation vm")
	}
	// skip the stack top value; it was already validated
	stack := vm.Stack()
	_, err = stack.Pop()
	if err != nil {
		return errors.Wrap(err, "getting top stack value")
	}
	var retainFor math.Duration
	if stack.Depth() > 0 {
		retainI, err := vm.Stack().PopAsInt64()
		if err != nil {
			return errors.Wrap(err, "getting retain duration from rules vm")
		}
		retainFor = math.Duration(retainI)
	}

	return app.UpdateState(
		app.applyTxDetails(tx),
		app.Unstake(tx.Qty, tx.Target, tx.StakeTo, tx.Rules, retainFor))
}

// GetSource implements Sourcer
func (tx *Unstake) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Unstake) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Unstake) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Unstake) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

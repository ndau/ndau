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
	"bytes"
	"fmt"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate returns nil if tx is valid, or an error
func (tx *CreateChildAccount) Validate(appI interface{}) error {
	// Ensure the target and child address are valid.
	_, err := address.Validate(tx.Target.String())
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("Target account address invalid: %s", tx.Target.String()))
	}
	_, err = address.Validate(tx.Child.String())
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("Child account address invalid: %s", tx.Child.String()))
	}

	// Verify that the child ownership key submitted generates the child address being created.
	childKind := tx.Child.Kind()
	if !address.IsValidKind(childKind) {
		return fmt.Errorf("Child account %s has invalid address kind: %x",
			tx.Child.String(), childKind)
	}
	childOwnershipAddress, err := address.Generate(childKind, tx.ChildOwnership.KeyBytes())
	if err != nil {
		return errors.Wrap(err, "Unable to generate address for child ownership key")
	}
	if tx.Child.String() != childOwnershipAddress.String() {
		return errors.New("Child ownership key and address do not match")
	}

	// The child signature should properly sign the address bytes using the child ownership key.
	if !tx.ChildSignature.Verify([]byte(tx.Child.String()), tx.ChildOwnership) {
		return errors.New("Invalid child signature")
	}

	// Similar to SetValidation tx: there must be at least 1 and no more than a const validation keys
	// keys set in this tx.
	if len(tx.ChildValidationKeys) < 1 || len(tx.ChildValidationKeys) > backing.MaxKeysInAccount {
		return fmt.Errorf("Expect between 1 and %d validation keys; got %d",
			backing.MaxKeysInAccount, len(tx.ChildValidationKeys))
	}

	// No child validation key may be equal to the child ownership key.
	for _, tk := range tx.ChildValidationKeys {
		if bytes.Equal(tk.KeyBytes(), tx.ChildOwnership.KeyBytes()) {
			return errors.New("Child ownership key may not be used as a validation key")
		}
	}

	// Ensure the validation scripts are chaincode.
	if len(tx.ChildValidationScript) > 0 && !IsChaincode(tx.ChildValidationScript) {
		return errors.New("Child validation script must be chaincode")
	}

	app := appI.(*App)

	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	child, _ := app.getAccount(tx.Child)

	// Ensure the child account does not already exist on the blockchain
	// This consequently also ensures that we won't have ancestry loops.
	// It also ensures that the child account does not yet have a Parent or Progenitor set,
	// as that happens in Apply().  Those are set precisely when its validation keys are set.
	// So only checking validation keys here is sufficient.
	if len(child.ValidationKeys) > 0 {
		return errors.New("child account may not already have validation keys")
	}

	// Below we ensure that the child account is not locked.
	// This rule only applies for exchange accounts.

	// We look at the target (i.e. parent) account for whether the child is (rather, will be) an
	// exchange account.  The parent account is fully established at this point; the child account
	// may not be.  In Apply(), the child account will be "adopted" and automatically inherit all
	// attributes from the parent when the child's Progenitor is set.
	isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(tx.Target, sv.AccountAttributeExchange)
	if err != nil {
		return err
	}

	// Ensure the child account is not locked, which could happen, for example, if a
	// TransferAndLock happened on the child prior to being created as a child account.
	if isExchangeAccount && child.IsLocked(app.BlockTime()) {
		return errors.New("Cannot create a locked child exchange account")
	}

	return nil
}

// Apply applies this tx if no error occurs
func (tx *CreateChildAccount) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), app.Delegate(tx.Child, tx.ChildDelegationNode), func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		child, _ := app.getAccount(tx.Child)
		child.ValidationKeys = tx.ChildValidationKeys
		child.ValidationScript = tx.ChildValidationScript

		target, _ := app.getAccount(tx.Target)
		child.Parent = &tx.Target
		if target.Progenitor == nil {
			child.Progenitor = &tx.Target
		} else {
			child.Progenitor = target.Progenitor
		}

		period := tx.ChildRecoursePeriod
		if period < 0 {
			period = app.getDefaultRecourseDuration()
		}
		child.RecourseSettings.Period = period

		st.Accounts[tx.Child.String()] = child

		return st, nil
	})
}

// GetSource implements Sourcer
func (tx *CreateChildAccount) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *CreateChildAccount) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *CreateChildAccount) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *CreateChildAccount) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *CreateChildAccount) GetAccountAddresses(app *App) ([]string, error) {
	return []string{tx.Target.String(), tx.Child.String()}, nil
}

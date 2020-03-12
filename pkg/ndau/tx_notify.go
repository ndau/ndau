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
	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Notify) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if !hasAccount {
		return errors.New("No such account")
	}

	if accountData.Lock == nil {
		return errors.New("Account is not locked")
	}
	if accountData.Lock.UnlocksOn != nil {
		return errors.New("Account has already been notified")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Notify) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := app.getAccount(tx.Target)

		uo := app.BlockTime().Add(accountData.Lock.NoticePeriod)
		accountData.Lock.UnlocksOn = &uo

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *Notify) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Notify) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Notify) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Notify) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

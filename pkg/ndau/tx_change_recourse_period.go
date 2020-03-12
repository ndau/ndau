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
func (tx *ChangeRecoursePeriod) Validate(appI interface{}) (err error) {
	app := appI.(*App)

	if tx.Period < 0 {
		return errors.New("Negative recourse period")
	}
	_, _, _, err = app.getTxAccount(tx)
	if err != nil {
		return err
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *ChangeRecoursePeriod) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		acct, _ := app.getAccount(tx.Target)

		ca := app.BlockTime().Add(acct.RecourseSettings.Period)
		acct.RecourseSettings.ChangesAt = &ca
		acct.RecourseSettings.Next = &tx.Period

		state.Accounts[tx.Target.String()] = acct
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *ChangeRecoursePeriod) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *ChangeRecoursePeriod) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ChangeRecoursePeriod) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ChangeRecoursePeriod) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

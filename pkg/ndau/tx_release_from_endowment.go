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
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *ReleaseFromEndowment) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("RFE qty may not be <= 0")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *ReleaseFromEndowment) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		var err error
		state := stateI.(*backing.State)

		acct, _ := app.getAccount(tx.Destination)
		acct.Balance, err = acct.Balance.Add(tx.Qty)
		if err != nil {
			return state, err
		}
		acct.UpdateCurrencySeat(app.BlockTime())
		state.Accounts[tx.Destination.String()] = acct

		// we give up overflow protection here in exchange for error-free
		// operation; we have external constraints that we will never issue
		// more than (30 million) * (100 million) napu, or 0.03% of 64-bits,
		// so this should be fine
		state.TotalRFE += tx.Qty

		// JSG the above might have modified total ndau in circulation, so recalculate SIB
		if app.IsFeatureActive("AllRFEInCirculation") {
			sib, target, err := app.calculateCurrentSIB(state, -1, -1)
			if err != nil {
				return state, err
			}
			state.SIB = sib
			state.TargetPrice = target
		}

		return state, err
	})
}

// GetSource implements Sourcer
func (tx *ReleaseFromEndowment) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ReleaseFromEndowmentAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *ReleaseFromEndowment) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *ReleaseFromEndowment) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *ReleaseFromEndowment) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *ReleaseFromEndowment) GetAccountAddresses(app *App) ([]string, error) {
	rfea, err := tx.GetSource(app)
	if err != nil {
		return nil, errors.Wrap(err, "getting RFE SV")
	}
	return []string{rfea.String(), tx.Destination.String()}, nil
}

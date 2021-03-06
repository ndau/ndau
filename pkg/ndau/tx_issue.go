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
	"github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Issue) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("Issue qty may not be <= 0")
	}

	state := app.GetState().(*backing.State)
	if state.TotalIssue+tx.Qty > state.TotalRFE {
		return errors.New("cannot issue more ndau than have been RFE'd")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *Issue) Apply(appI interface{}) error {
	app := appI.(*App)

	return app.UpdateState(
		app.applyTxDetails(tx),
		func(stateI metast.State) (metast.State, error) {
			state := stateI.(*backing.State)

			// we give up overflow protection here in exchange for error-free
			// operation; we have external constraints that we will never issue
			// more than (30 million) * (100 million) napu, or 0.03% of 64-bits,
			// so this should be fine
			state.TotalIssue += tx.Qty
			return state, nil
		},
		app.updatePricesAndSIB(-1),
	)
}

// GetSource implements Sourcer
func (tx *Issue) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.ReleaseFromEndowmentAddressName, &addr)
	return
}

// GetSequence implements Sequencer
func (tx *Issue) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Issue) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Issue) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// UpdatedTargetPrice implements search.TargetPriceIndexable
func (*Issue) UpdatedTargetPrice() {}

var _ search.TargetPriceIndexable = (*Issue)(nil)

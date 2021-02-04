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
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate satisfies metatx.Transactable
func (tx *Transfer) Validate(appInt interface{}) error {
	app := appInt.(*App)

	if tx.Qty <= math.Ndau(0) {
		return errors.New("invalid transfer: Qty not positive")
	}

	if tx.Source == tx.Destination {
		return errors.New("invalid transfer: source == destination")
	}

	source, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if source.IsLocked(app.BlockTime()) {
		return errors.New("source is locked")
	}

	dest, _ := app.getAccount(tx.Destination)

	if dest.IsNotified(app.BlockTime()) {
		return errors.New("transfers into notified addresses are invalid")
	}

	return nil
}

// Apply satisfies metatx.Transactable
func (tx *Transfer) Apply(appInt interface{}) error {
	app := appInt.(*App)

	var updater func(...func(metast.State) (metast.State, error)) error
	updater = app.UpdateState
	if !app.IsFeatureActive("NoLeakyUpdateState") {
		updater = app.UpdateStateLeaky
	}

	return updater(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		source, _ := app.getAccount(tx.Source)
		dest, _ := app.getAccount(tx.Destination)

		err := (&dest.WeightedAverageAge).UpdateWeightedAverageAge(
			app.BlockTime().Since(dest.LastWAAUpdate),
			tx.Qty,
			dest.Balance,
		)
		if err != nil {
			return stateI, errors.Wrap(err, "update waa")
		}
		dest.LastWAAUpdate = app.BlockTime()

		dest.Balance += tx.Qty

		destIsExchange := false
		if app.IsFeatureActive("NoExchangeHoldsOnTransfer") {
			destIsExchange, err = app.GetState().(*backing.State).AccountHasAttribute(tx.Destination, sv.AccountAttributeExchange)
			if err != nil {
				return stateI, errors.New("dest account exchange attribute can't be retrieved")
			}
		}

		if source.RecourseSettings.Period != 0 && !destIsExchange {
			x := app.BlockTime().Add(source.RecourseSettings.Period)
			dest.Holds = append(dest.Holds, backing.Hold{
				Qty:    tx.Qty,
				Expiry: &x,
				Txhash: metatx.Hash(tx),
			})
		}

		dest.UpdateCurrencySeat(app.BlockTime())

		state := stateI.(*backing.State)

		state.Accounts[tx.Destination.String()] = dest
		state.Accounts[tx.Source.String()] = source

		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *Transfer) GetSource(*App) (address.Address, error) {
	return tx.Source, nil
}

// GetDestination implements HasDestination
func (tx *Transfer) GetDestination(*App) (address.Address, error) {
	return tx.Destination, nil
}

// Withdrawal implements Withdrawer
func (tx *Transfer) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements Sequencer
func (tx *Transfer) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Transfer) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Transfer) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *Transfer) GetAccountAddresses(app *App) ([]string, error) {
	return []string{tx.Source.String(), tx.Destination.String()}, nil
}

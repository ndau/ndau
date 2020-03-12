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
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *Lock) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if !hasAccount {
		return errors.New("No such account")
	}

	if accountData.Lock != nil {
		if accountData.Lock.UnlocksOn == nil {
			// if not notified, lock is valid if its period >= the current period
			if tx.Period < accountData.Lock.NoticePeriod {
				return errors.New("Locked, non-notified accounts may be relocked only with periods >= their current")
			}
		} else {
			// if notified, lock is valid if it expires after the current unlock time
			lockExpiry := app.BlockTime().Add(tx.Period)
			uo := *accountData.Lock.UnlocksOn
			if lockExpiry.Compare(uo) < 0 {
				return errors.New("Locked, notified accounts may be relocked only when new lock min expiry >= current unlock time")
			}
		}
	}

	// Ensure that this is not an exchange account, as they are not allowed to be locked.
	isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(tx.Target, sv.AccountAttributeExchange)
	if err != nil {
		return err
	}
	if isExchangeAccount {
		return errors.New("Cannot lock exchange accounts")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Lock) Apply(appI interface{}) error {
	app := appI.(*App)

	lockedBonusRateTable := eai.RateTable{}
	err := app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := app.getAccount(tx.Target)

		accountData.Lock = backing.NewLock(tx.Period, lockedBonusRateTable)

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *Lock) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *Lock) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Lock) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Lock) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

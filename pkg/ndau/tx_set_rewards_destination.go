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
	"github.com/pkg/errors"
)

// Validate implements metatx.Transactable
func (tx *SetRewardsDestination) Validate(appI interface{}) error {
	app := appI.(*App)

	accountData, hasAccount, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}
	if !hasAccount {
		return errors.New("No such account")
	}
	_, err = address.Validate(tx.Destination.String())
	if err != nil {
		return errors.Wrap(err, "Destination")
	}

	// source account must not be receiving rewards from any other account
	if len(accountData.IncomingRewardsFrom) > 0 {
		return fmt.Errorf("accounts may not both send and receive rewards. Source receives rewards from these accounts: %s", accountData.IncomingRewardsFrom)
	}

	// if dest is the same as source, we're resetting the EAI to accumulate
	// in its account of origin.
	// neither destination rule appllies in that case.
	if tx.Destination.String() != tx.Target.String() {
		// dest account must not be sending rewards to any other account
		targetData, _ := app.getAccount(tx.Destination)
		if targetData.RewardsTarget != nil {
			return fmt.Errorf("accounts may not both send and receive rewards. Destination sends rewards to %s", *targetData.RewardsTarget)
		}

		// dest account must not be notified
		if targetData.IsNotified(app.BlockTime()) {
			return errors.New("Destination is currently notified and may not receive new rewards until it unlocks")
		}
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *SetRewardsDestination) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		accountData, _ := app.getAccount(tx.Target)

		targetData, _ := app.getAccount(tx.Destination)

		// update inbound of rewards target
		if accountData.RewardsTarget != nil && accountData.RewardsTarget.String() != tx.Target.String() {
			oldTargetData, _ := app.getAccount(*accountData.RewardsTarget)
			// remove account from current target inbounds list
			inbounds := make([]address.Address, 0, len(oldTargetData.IncomingRewardsFrom)-1)
			for _, addr := range oldTargetData.IncomingRewardsFrom {
				if tx.Target.String() != addr.String() {
					inbounds = append(inbounds, addr)
				}
			}
			oldTargetData.IncomingRewardsFrom = inbounds
			state.Accounts[accountData.RewardsTarget.String()] = oldTargetData
		}

		if tx.Target.String() == tx.Destination.String() {
			accountData.RewardsTarget = nil
		} else {
			accountData.RewardsTarget = &tx.Destination
			targetData.IncomingRewardsFrom = append(targetData.IncomingRewardsFrom, tx.Target)
			state.Accounts[tx.Destination.String()] = targetData
		}

		state.Accounts[tx.Target.String()] = accountData
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *SetRewardsDestination) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// GetSequence implements Sequencer
func (tx *SetRewardsDestination) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *SetRewardsDestination) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SetRewardsDestination) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetAccountAddresses returns the account addresses associated with this transaction type.
func (tx *SetRewardsDestination) GetAccountAddresses(app *App) ([]string, error) {
	return []string{tx.Target.String(), tx.Destination.String()}, nil
}

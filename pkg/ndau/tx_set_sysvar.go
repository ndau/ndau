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
	"encoding/base64"
	"fmt"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Validate implements metatx.Transactable
func (tx *SetSysvar) Validate(appI interface{}) error {
	app := appI.(*App)

	if app.IsFeatureActive("SysvarValidityCheck") {
		validity := sv.IsValid(tx.Name, tx.Value)
		if validity == nil {
			app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"sysvar.name": tx.Name,
			}).Warn("sysvar has no validation configured")
		}
		if validity != nil && !*validity {
			app.DecoratedTxLogger(tx).WithFields(log.Fields{
				"sysvar.name":  tx.Name,
				"sysvar.value": base64.StdEncoding.EncodeToString(tx.Value),
			}).Info("rejected sysvar: failed validation")
			return errors.New("sysvar validation failed")
		}
	}

	// if we let someone overwrite the sysvar governing who is allowed to
	// set the sysvar with bad data, then we're hosed. Let's ensure that
	// if that's the sysvar being set, it's by an account which has been
	// exists and has at least one validation key.
	if tx.Name == sv.SetSysvarAddressName {
		var acct address.Address
		leftovers, err := acct.UnmarshalMsg(tx.Value)
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf(
					"value for %s must be a valid Address",
					sv.SetSysvarAddressName,
				),
			)
		}
		if len(leftovers) > 0 {
			return fmt.Errorf(
				"value for %s must not have leftovers; got %x",
				sv.SetSysvarAddressName,
				leftovers,
			)
		}

		data, exists := app.getAccount(acct)
		if !exists {
			return fmt.Errorf(
				"new %s must be an account which exists; %s doesn't",
				sv.SetSysvarAddressName,
				acct,
			)
		}

		if len(data.ValidationKeys) == 0 {
			return fmt.Errorf(
				"new %s acct (%s) must have at least 1 validation key",
				sv.SetSysvarAddressName,
				acct,
			)
		}
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *SetSysvar) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		state.Sysvars[tx.Name] = tx.Value

		// JSG the above might have modified the SIB script, so recalculate SIB
		if app.IsFeatureActive("AllRFEInCirculation") && tx.Name == sv.SIBScriptName {
			sib, target, err := app.calculateCurrentSIB(state, -1, -1)
			if err != nil {
				return state, err
			}
			state.SIB = sib
			state.TargetPrice = target
		}
		return state, nil
	})
}

// GetSource implements Sourcer
func (tx *SetSysvar) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.SetSysvarAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not properly set; SetSysvar therefore disallowed",
			sv.SetSysvarAddressName,
		)
		return
	}
	return
}

// GetSequence implements Sequencer
func (tx *SetSysvar) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *SetSysvar) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *SetSysvar) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetName implements SysvarIndexable.
func (tx *SetSysvar) GetName() string {
	return tx.Name
}

// GetValue implements SysvarIndexable.
func (tx *SetSysvar) GetValue() []byte {
	return tx.Value
}

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
	"errors"

	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	log "github.com/sirupsen/logrus"
)

// Validate implements metatx.Transactable
func (tx *Burn) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("burn qty must be positive")
	}

	acctData, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if acctData.IsLocked(app.BlockTime()) {
		return errors.New("burn from locked accounts prohibited")
	}

	if len(tx.EthAddr) > 0 {
		logger := app.DecoratedLogger()
		logger.WithFields(log.Fields{
			"EthAddr": tx.EthAddr,
			"Target":  tx.Target,
		}).Info("Burn with optional EthAddr")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *Burn) Apply(appI interface{}) error {
	app := appI.(*App)

	lockedBonusRateTable := eai.RateTable{}
	err := app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	return app.UpdateState(app.applyTxDetails(tx), func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.TotalBurned += tx.Qty

		// JSG the above might have modified total ndau in circulation, so recalculate SIB
		if app.IsFeatureActive("AllRFEInCirculation") {
			sib, target, err := app.calculateCurrentSIB(st, -1, -1)
			if err != nil {
				return st, err
			}
			st.SIB = sib
			st.TargetPrice = target
		}

		return st, nil
	})
}

// GetSource implements Sourcer
func (tx *Burn) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// Withdrawal implements Withdrawer
func (tx *Burn) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements Sequencer
func (tx *Burn) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *Burn) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *Burn) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

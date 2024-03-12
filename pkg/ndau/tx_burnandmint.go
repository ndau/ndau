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
	"encoding/hex"
	"errors"
	"math/big"

	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
)

// Validate implements metatx.Transactable
func (tx *BurnAndMint) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.Qty <= 0 {
		return errors.New("burn qty must be positive")
	}

	// TODO: Improve checking for valid Ethereum address

	if tx.EthAddr == "" {
		return errors.New("Ethereum address may not be empty")
	}

	acctData, _, _, err := app.getTxAccount(tx)
	if err != nil {
		return err
	}

	if acctData.IsLocked(app.BlockTime()) {
		return errors.New("burn from locked accounts prohibited")
	}

	return nil
}

// Apply implements metatx.Transactable
func (tx *BurnAndMint) Apply(appI interface{}) error {
	app := appI.(*App)

	lockedBonusRateTable := eai.RateTable{}
	err := app.System(sv.LockedRateTableName, &lockedBonusRateTable)
	if err != nil {
		return err
	}

	// Send minting vote to the NPAY smart contract.
	// var hash [32]byte
	hash, _ := hex.DecodeString(metatx.Hash(tx))
	var txHash [32]byte
	copy(txHash[:], hash[:32])
	MintNPAY(txHash, big.NewInt(987654), big.NewInt(int64(tx.Qty)), tx.EthAddr)

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
func (tx *BurnAndMint) GetSource(*App) (address.Address, error) {
	return tx.Target, nil
}

// Withdrawal implements Withdrawer
func (tx *BurnAndMint) Withdrawal() math.Ndau {
	return tx.Qty
}

// GetSequence implements Sequencer
func (tx *BurnAndMint) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *BurnAndMint) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *BurnAndMint) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

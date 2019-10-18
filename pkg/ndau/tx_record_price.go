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

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/signed"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
)

// return a function intended to be run within app.UpdateState
//
// special case: if the input is negative, just use the existing value
func (app *App) updatePricesAndSIB(marketPrice pricecurve.Nanocent) func(stateI metast.State) (metast.State, error) {
	return func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)
		if marketPrice < 0 {
			marketPrice = state.MarketPrice
		}
		sib, target, err := app.calculateCurrentSIB(state, marketPrice, -1)
		if err != nil {
			return stateI, err
		}
		state.SIB = sib
		state.MarketPrice = marketPrice
		state.TargetPrice = target

		return state, err
	}
}

func floorPrice(app *App, nav pricecurve.Nanocent) (fp pricecurve.Nanocent, err error) {
	summary := getLastSummary(app)
	// just dividing NAV (denominated in nanocents) by TotalCirculation (denominated
	// in napu) gives us nanocents per napu, which is inconsistent with market
	// price and target price (nanocents per ndau), and also small enough that
	// we'd likely see non-trivial rounding errors. Using a muldiv here gives us
	// nanocents per ndau without overflow.

	// default zero: avoid divide by 0 errors
	if summary.TotalCirculation != 0 {
		var floorPriceI int64
		floorPriceI, err = signed.MulDiv(
			int64(nav),
			constants.NapuPerNdau,
			int64(summary.TotalCirculation*2))
		if err != nil {
			err = errors.Wrap(err, "computing floor price")
			return
		}
		fp = pricecurve.Nanocent(floorPriceI)
	}
	return
}

// calculates the SIB implied by the market price given the current app state.
//
// It also returns the calculated target price.
func (app *App) calculateCurrentSIB(state *backing.State, marketPrice, nav pricecurve.Nanocent) (sib eai.Rate, targetPrice pricecurve.Nanocent, err error) {
	if marketPrice < 0 {
		marketPrice = state.MarketPrice
	}
	if nav < 0 {
		nav = state.GetEndowmentNAV()
	}

	// compute the current target price
	if app.IsFeatureActive("TargetPrice9999") {
		targetPrice, err = pricecurve.PriceAtUnit9999(state.TotalIssue)
	} else {
		targetPrice, err = pricecurve.PriceAtUnit10000(state.TotalIssue)
	}
	if err != nil {
		err = errors.Wrap(err, "computing target price")
		return
	}

	var fp pricecurve.Nanocent
	fp, err = floorPrice(app, nav)
	if err != nil {
		err = errors.Wrap(err, "computing floor price")
		return
	}

	// get the script used to perform the calculation
	var sibScript wkt.Bytes
	err = app.System(sv.SIBScriptName, &sibScript)
	if err != nil {
		err = errors.Wrap(err, "fetching "+sv.SIBScriptName)
		return
	}
	if !IsChaincode(sibScript) {
		err = errors.New("sibScript appears not to be chaincode")
		return
	}

	// compute SIB
	vm, err := BuildVMForSIB(sibScript, uint64(targetPrice), uint64(marketPrice), uint64(fp), app.BlockTime())
	if err != nil {
		err = errors.Wrap(err, "building vm for SIB calculation")
		return
	}

	err = vm.Run(nil)
	if err != nil {
		err = errors.Wrap(err, "computing SIB")
		return
	}

	top, err := vm.Stack().PopAsInt64()
	if err != nil {
		err = errors.Wrap(err, "retrieving SIB from VM")
		return
	}

	sib = eai.Rate(top)
	return
}

// Validate implements metatx.Transactable
func (tx *RecordPrice) Validate(appI interface{}) error {
	app := appI.(*App)

	if tx.MarketPrice <= 0 {
		return errors.New("RecordPrice market price may not be <= 0")
	}

	_, _, _, err := app.getTxAccount(tx)

	return err
}

// Apply implements metatx.Transactable
func (tx *RecordPrice) Apply(appI interface{}) error {
	app := appI.(*App)
	return app.UpdateState(app.applyTxDetails(tx), app.updatePricesAndSIB(tx.MarketPrice))
}

// GetSource implements Sourcer
func (tx *RecordPrice) GetSource(app *App) (addr address.Address, err error) {
	err = app.System(sv.RecordPriceAddressName, &addr)
	if err != nil {
		return
	}
	if addr.Revalidate() != nil {
		err = fmt.Errorf(
			"%s sysvar not set; RecordPrice therefore disallowed",
			sv.RecordPriceAddressName,
		)
		return
	}
	return
}

// GetSequence implements Sequencer
func (tx *RecordPrice) GetSequence() uint64 {
	return tx.Sequence
}

// GetSignatures implements Signeder
func (tx *RecordPrice) GetSignatures() []signature.Signature {
	return tx.Signatures
}

// ExtendSignatures implements Signable
func (tx *RecordPrice) ExtendSignatures(sa []signature.Signature) {
	tx.Signatures = append(tx.Signatures, sa...)
}

// GetMarketPrice implements search.MarketPriceIndexable
func (tx *RecordPrice) GetMarketPrice() pricecurve.Nanocent {
	return tx.MarketPrice
}

var _ search.MarketPriceIndexable = (*RecordPrice)(nil)

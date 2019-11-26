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
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const recordEndowmentNAVKeys = "recordEndowmentNAV private keys"

func initAppRecordEndowmentNAV(t *testing.T) (*App, generator.Associated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// fetch the RecordEndowmentNAV address system variable
	recordEndowmentNAVAddr := address.Address{}
	err := app.System(sv.RecordEndowmentNAVAddressName, &recordEndowmentNAVAddr)
	require.NoError(t, err)
	assc[recordEndowmentNAVKeys], err = MockSystemAccount(app, recordEndowmentNAVAddr)

	// special acct contains exactly 1 ndau
	modify(t, recordEndowmentNAVAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1 * constants.NapuPerNdau
	})

	// reset the last summary so we recalculate on each test run
	lastSummary.BlockHeight = ^uint64(0)

	return app, assc
}

func TestRecordEndowmentNAVIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRecordEndowmentNAV(t)
	privateKeys := assc[recordEndowmentNAVKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			recordEndowmentNAV := NewRecordEndowmentNAV(
				pricecurve.Nanocent(1),
				1,
				private,
			)

			recordEndowmentNAVBytes, err := tx.Marshal(recordEndowmentNAV, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: recordEndowmentNAVBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestRecordEndowmentNAVIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRecordEndowmentNAV(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	recordEndowmentNAV := NewRecordEndowmentNAV(
		pricecurve.Nanocent(1),
		1,
		private,
	)

	recordEndowmentNAVBytes, err := tx.Marshal(recordEndowmentNAV, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: recordEndowmentNAVBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRecordEndowmentNAVIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppRecordEndowmentNAV(t)
	privateKeys := assc[recordEndowmentNAVKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.RecordEndowmentNAVAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 recordEndowmentNAV keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			recordEndowmentNAV := NewRecordEndowmentNAV(
				pricecurve.Nanocent(1),
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, recordEndowmentNAV)

			var expect code.ReturnCode
			if i == 0 {
				expect = code.OK
			} else {
				expect = code.InvalidTransaction
			}
			require.Equal(t, expect, code.ReturnCode(resp.Code))
		})
	}
}

func TestNAVMustBePositive(t *testing.T) {
	app, assc := initAppRecordEndowmentNAV(t)
	privateKeys := assc[recordEndowmentNAVKeys].([]signature.PrivateKey)

	recordEndowmentNAV := NewRecordEndowmentNAV(
		pricecurve.Nanocent(-1),
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, recordEndowmentNAV)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	require.Contains(t, resp.Log, "NAV")
	require.Contains(t, resp.Log, "may not be <= 0")
}

func TestNAVMustChangeSIB(t *testing.T) {
	app, assc := initAppRecordEndowmentNAV(t)
	privateKeys := assc[recordEndowmentNAVKeys].([]signature.PrivateKey)

	// set nonzero market price and nav
	const (
		issue      = 5 * 1000000 * constants.NapuPerNdau
		dollar     = 100000000000 // nanocent
		market     = 12 * dollar  // intentionally low, to incur SIB
		initialNAV = 15 * 1000000 * dollar
		finalNAV   = 20 * 1000000 * dollar
	)
	var initialSIB eai.Rate
	var err error
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.TotalIssue = issue
		var target pricecurve.Nanocent
		initialSIB, target, err = app.calculateCurrentSIB(state, market, initialNAV)
		require.NoError(t, err)
		state.SIB = initialSIB
		state.MarketPrice = market
		state.TargetPrice = target
		state.SetEndowmentNAV(initialNAV)
		return state, nil
	})
	require.NoError(t, err)
	require.NotZero(t, initialSIB)

	recordEndowmentNAV := NewRecordEndowmentNAV(
		finalNAV,
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, recordEndowmentNAV)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	require.NotZero(t, app.GetState().(*backing.State).SIB)
	require.NotEqual(t, initialSIB, app.GetState().(*backing.State).SIB)
}

func TestNAVChangesState(t *testing.T) {
	app, assc := initAppRecordEndowmentNAV(t)
	privateKeys := assc[recordEndowmentNAVKeys].([]signature.PrivateKey)

	// set nonzero market price and nav
	const (
		issue      = 5 * 1000000 * constants.NapuPerNdau
		dollar     = 100000000000 // nanocent
		market     = 12 * dollar  // intentionally low, to incur SIB
		initialNAV = 15 * 1000000 * dollar
		finalNAV   = 20 * 1000000 * dollar
	)

	state := app.GetState().(*backing.State)
	require.False(t, state.IsManagedVarSet("EndowmentNAV"))
	require.Equal(t, pricecurve.Nanocent(0), state.GetEndowmentNAV())

	err := app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.TotalIssue = issue
		state.MarketPrice = market

		require.False(t, state.IsManagedVarSet("EndowmentNAV"))
		require.Equal(t, pricecurve.Nanocent(0), state.GetEndowmentNAV())
		state.SetEndowmentNAV(initialNAV)
		require.True(t, state.IsManagedVarSet("EndowmentNAV"))
		require.Equal(t, pricecurve.Nanocent(initialNAV), state.GetEndowmentNAV())

		return state, nil
	})
	require.NoError(t, err)

	recordEndowmentNAV := NewRecordEndowmentNAV(
		finalNAV,
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, recordEndowmentNAV)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	state = app.GetState().(*backing.State)
	require.True(t, state.IsManagedVarSet("EndowmentNAV"))
	require.Equal(t, pricecurve.Nanocent(finalNAV), state.GetEndowmentNAV())
}

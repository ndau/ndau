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

	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/pricecurve"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/ndau/metanode/pkg/meta/app/code"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
)

func TestIssueIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			issue := NewIssue(
				math.Ndau(1),
				1,
				private,
			)

			rfeBytes, err := tx.Marshal(issue, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestIssueIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRFE(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)

	issue := NewIssue(
		math.Ndau(1),
		1,
		private,
	)

	rfeBytes, err := tx.Marshal(issue, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidIssueAddsNdauToTotal(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			oldIssuance := app.GetState().(*backing.State).TotalIssue

			issue := NewIssue(
				math.Ndau(1),
				uint64(i+1),
				private,
			)

			resp := deliverTx(t, app, issue)
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
			require.Equal(t, oldIssuance+1, app.GetState().(*backing.State).TotalIssue)
		})
	}
}

func TestCannotIssueMoreThanRFEd(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			oldIssuance := app.GetState().(*backing.State).TotalIssue

			issue := NewIssue(
				math.Ndau(101*constants.NapuPerNdau),
				uint64(i+1),
				private,
			)

			resp := deliverTx(t, app, issue)
			require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
			require.Equal(t, oldIssuance, app.GetState().(*backing.State).TotalIssue)
		})
	}
}

func TestIssueIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.ReleaseFromEndowmentAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 rfe keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			issue := NewIssue(
				math.Ndau(1),
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, issue)

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

func TestIssueDoesntAdjustMarketPrice(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	oldMarketPrice := app.GetState().(*backing.State).MarketPrice

	issue := NewIssue(
		math.Ndau(1),
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, issue)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, oldMarketPrice, app.GetState().(*backing.State).MarketPrice)
}

func TestIssueAdjustsTargetAndSIB(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	getVars := func() (pricecurve.Nanocent, eai.Rate) {
		state := app.GetState().(*backing.State)
		return state.TargetPrice, state.SIB
	}
	oldTargetPrice, oldSIB := getVars()

	issue := NewIssue(
		math.Ndau(1),
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, issue)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	targetPrice, sib := getVars()
	require.NotEqual(t, oldTargetPrice, targetPrice)
	require.NotEqual(t, oldSIB, sib)
}

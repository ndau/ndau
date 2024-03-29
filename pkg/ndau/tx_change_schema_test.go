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
	"fmt"
	"testing"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	generator "github.com/ndau/system_vars/pkg/genesis.generator"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const changeSchemaKeys = "changeSchema private keys"

var hasQuit bool

func initAppChangeSchema(t *testing.T) (*App, generator.Associated) {
	return initAppChangeSchemaWithIndex(t, "", -1)
}

func getChangeSchemaAddr(t *testing.T, app *App) (addr address.Address) {
	// fetch the ChangeSchema address system variable
	err := app.System(sv.ChangeSchemaAddressName, &addr)
	require.NoError(t, err)
	return
}

func initAppChangeSchemaWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	*App, generator.Associated,
) {
	app, assc := initAppWithIndex(t, indexAddr, indexVersion)
	app.InitChain(abci.RequestInitChain{})

	// fetch the ChangeSchema address system variable
	changeSchemaAddr := getChangeSchemaAddr(t, app)
	var err error
	assc[changeSchemaKeys], err = MockSystemAccount(app, changeSchemaAddr)
	require.NoError(t, err)

	// ensure special acct contains exactly 1 napu so balance test works
	modify(t, changeSchemaAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// replace quit helper so it doesn't actually exit the test
	hasQuit = false
	quit = func() {
		hasQuit = true
		app.SetStateValidity(errors.New("if we quit before commit, subsequent txs in the block fail"))
	}

	return app, assc
}

func TestChangeSchemaIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppChangeSchema(t)
	privateKeys := assc[changeSchemaKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			changeSchema := NewChangeSchema(
				"",
				1,
				private,
			)

			changeSchemaBytes, err := metatx.Marshal(changeSchema, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: changeSchemaBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestChangeSchemaIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppChangeSchema(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	changeSchema := NewChangeSchema(
		"",
		1,
		private,
	)

	changeSchemaBytes, err := metatx.Marshal(changeSchema, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: changeSchemaBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestChangeSchemaIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppChangeSchema(t)
	privateKeys := assc[changeSchemaKeys].([]signature.PrivateKey)

	// this test only works if we don't actually invalidate the app state
	// on quit
	quit = func() {}

	txFeeAddr := address.Address{}
	err := app.System(sv.ReleaseFromEndowmentAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 changeSchema keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			changeSchema := NewChangeSchema(
				"",
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, changeSchema)

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

func TestChangeSchemaCallsQuitFunction(t *testing.T) {
	app, assc := initAppChangeSchema(t)
	privateKeys := assc[changeSchemaKeys].([]signature.PrivateKey)

	changeSchema := NewChangeSchema(
		"",
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, changeSchema)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// quit only calls at the beginning of the next block
	require.False(t, hasQuit)

	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	subsequent := NewChangeValidation(getChangeSchemaAddr(t, app), []signature.PublicKey{public}, nil, 2, private)
	resp = deliverTx(t, app, subsequent)
	require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
	require.True(t, hasQuit)
}

func TestChangeSchemaCallsQuitFunctionAfterNomsCommit(t *testing.T) {
	app, assc := initAppChangeSchema(t)
	privateKeys := assc[changeSchemaKeys].([]signature.PrivateKey)
	csAddr := getChangeSchemaAddr(t, app)
	csAcct, _ := app.getAccount(csAddr)

	changeSchema := NewChangeSchema(
		"",
		1,
		privateKeys...,
	)

	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	subsequent := NewChangeValidation(csAddr, []signature.PublicKey{public}, nil, 2, privateKeys...)

	resps, _ := deliverTxsContext(t, app, []metatx.Transactable{changeSchema, subsequent}, ddc(t))
	for _, resp := range resps {
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}
	// we shouldn't quit until the beginning of the subsequent block
	require.False(t, hasQuit)

	subsequent = NewChangeValidation(csAddr, csAcct.ValidationKeys, nil, 3, private)
	resp := deliverTx(t, app, subsequent)
	require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
	require.True(t, hasQuit)
}

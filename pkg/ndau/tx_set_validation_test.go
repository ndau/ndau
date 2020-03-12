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
	"time"

	"github.com/ndau/metanode/pkg/meta/app/code"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndau/pkg/query"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppSetValidation(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	return app
}

func TestSetValidationAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// the address is invalid, but NewSetValidation doesn't validate this
	ca := NewSetValidation(addr, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppSetValidation(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	ca = NewSetValidation(fakeTarget, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err = tx.Marshal(ca, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidSetValidation(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppSetValidation(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestSetValidationNewValidationKeyNotEqualOwnershipKey(t *testing.T) {
	app := initAppSetValidation(t)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{targetPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidSetValidationUpdatesValidationKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppSetValidation(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// apply the transaction
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	dresp := app.DeliverTx(abci.RequestDeliverTx{Tx: ctkBytes})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	t.Log(dresp.Log)
	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.KeyBytes(), ad.ValidationKeys[0].KeyBytes())
	})
}

func TestSetValidationNoValidationKeys(t *testing.T) {
	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppSetValidation(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetValidationTooManyValidationKeys(t *testing.T) {
	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	ca := NewSetValidation(targetAddress, targetPublic, newKeys, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppSetValidation(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetValidationCannotOverwriteMoreThanOneValidationKey(t *testing.T) {
	app := initAppSetValidation(t)

	existing1, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	existing2, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = []signature.PublicKey{existing1, existing2}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetValidationDeductsTxFee(t *testing.T) {
	app := initAppSetValidation(t)
	ensureRecent(t, app, targetAddress.String())
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		newPublic, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		tx := NewSetValidation(
			targetAddress,
			targetPublic,
			[]signature.PublicKey{newPublic},
			[]byte{},
			1+uint64(i),
			targetPrivate,
		)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestSetValidationDoesntResetWAA(t *testing.T) {
	// inspired by a Real Live Bug!
	app := initAppSetValidation(t)

	assertExistsAndNonzeroWAAUpdate := func(expectExists bool) {
		resp := app.Query(abci.RequestQuery{
			Path: query.AccountEndpoint,
			Data: []byte(targetAddress.String()),
		})
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		require.Equal(t, fmt.Sprintf(query.AccountInfoFmt, expectExists), resp.Info)

		accountData := new(backing.AccountData)
		_, err := accountData.UnmarshalMsg(resp.Value)
		require.NoError(t, err)

		require.NotZero(t, accountData.LastWAAUpdate)
	}

	assertExistsAndNonzeroWAAUpdate(false)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	resp := deliverTx(t, app, ca)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	assertExistsAndNonzeroWAAUpdate(true)
}

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
	"testing"

	"github.com/ndau/metanode/pkg/meta/app/code"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

func TestCreateChildAccountInvalidTargetAddress(t *testing.T) {
	app, private := initAppTx(t)

	// Flip the bits of the last byte so the address is no longer correct.
	addrBytes := []byte(sourceAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// Ensure that we didn't accidentally create a valid address.
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// The address is invalid, but NewCreateChildAccount doesn't validate this.
	cca := NewCreateChildAccount(
		addr,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	// However, the resultant transaction must not be valid.
	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateChildAccountInvalidChildAddress(t *testing.T) {
	app, private := initAppTx(t)

	// Flip the bits of the last byte so the address is no longer correct.
	addrBytes := []byte(sourceAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// Ensure that we didn't accidentally create a valid address.
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// The address is invalid, but NewCreateChildAccount doesn't validate this.
	cca := NewCreateChildAccount(
		sourceAddress,
		addr,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	// However, the resultant transaction must not be valid.
	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateChildAccountNonExistentTargetAddress(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		targetAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidCreateChildAccount(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	dresp := deliverTx(t, app, cca)
	t.Log(dresp.Log)

	// Ensure the child's recourse period matches the default from the system variable.
	child, _ := app.getAccount(childAddress)
	require.Equal(t, app.getDefaultRecourseDuration(), child.RecourseSettings.Period)
	// ensure we updated the delegation node
	require.NotNil(t, child.DelegationNode)
	require.Equal(t, cca.ChildDelegationNode, *child.DelegationNode)
}

func TestCreateChildAccountRecoursePeriod(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	period := math.Duration(1234)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		period,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	dresp := deliverTx(t, app, cca)
	t.Log(dresp.Log)

	// Ensure the child's recourse period matches what we set it to.
	child, _ := app.getAccount(childAddress)
	require.Equal(t, period, child.RecourseSettings.Period)
}

func TestCreateChildAccountNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app, private := initAppTx(t)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{childPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidCreateChildAccountUpdatesTransferKey(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	dresp := deliverTx(t, app, cca)
	t.Log(dresp.Log)

	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))
	modify(t, childAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.KeyBytes(), ad.ValidationKeys[0].KeyBytes())
	})
}

func TestCreateChildAccountNoValidationKeys(t *testing.T) {
	app, private := initAppTx(t)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateChildAccountTooManyValidationKeys(t *testing.T) {
	app, private := initAppTx(t)

	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		newKeys,
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateChildAccountCannotHappenTwice(t *testing.T) {
	app, private := initAppTx(t)

	// Simulate the child account already having been created.
	existing, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	modify(t, childAddress.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = []signature.PublicKey{existing}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateGrandchildAccount(t *testing.T) {
	app, sourceValidation := initAppTx(t)

	createChild := func(
		parent address.Address,
		progenitor address.Address,
		parentPrivate signature.PrivateKey,
	) (address.Address, signature.PrivateKey) {
		childPublic, childPrivate, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		child, err := address.Generate(parent.Kind(), childPublic.KeyBytes())
		require.NoError(t, err)

		childSignature := childPrivate.Sign([]byte(child.String()))

		validationPublic, validationPrivate, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		parentAcct, _ := app.getAccount(parent)

		cca := NewCreateChildAccount(
			parent,
			child,
			childPublic,
			childSignature,
			childRecoursePeriod,
			[]signature.PublicKey{validationPublic},
			[]byte{},
			child,
			parentAcct.Sequence+1,
			parentPrivate,
		)

		// Make the progenitor an exchange account, to test more code paths.
		context := ddc(t).withExchangeAccount(progenitor)
		dresp, _ := deliverTxContext(t, app, cca, context)
		require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

		childAcct, exists := app.getAccount(child)
		require.True(t, exists)
		require.Equal(t, &parent, childAcct.Parent)
		require.Equal(t, &progenitor, childAcct.Progenitor)
		require.Equal(t, 1, len(cca.ChildValidationKeys))
		require.Equal(t, 1, len(childAcct.ValidationKeys))
		require.Equal(t,
			cca.ChildValidationKeys[0].KeyBytes(),
			childAcct.ValidationKeys[0].KeyBytes(),
		)

		// Since the progenitor was marked as an exchange account, so should
		// any descendant.
		// However, whether or not any account is an exchange account depends
		// on the state of the sysvars; these have been changed by our context.
		// We therefore need to get into that context.
		context.Within(app, func() {
			isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(child, sv.AccountAttributeExchange)
			require.NoError(t, err)
			require.True(t, isExchangeAccount)
		})

		return child, validationPrivate
	}

	// Create a child of the source account.
	child, childValidation := createChild(sourceAddress, sourceAddress, sourceValidation)

	// Create a child of the child (a grandchild of the source account).
	createChild(child, sourceAddress, childValidation)
}

func TestCreateChildAccountInvalidValidationScript(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{0x01},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreateChildAccountInvalidChildSignature(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		private.Sign([]byte(childAddress.String())),
		childRecoursePeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

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

	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	generator "github.com/ndau/system_vars/pkg/genesis.generator"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const cvcKeys = "cvc private keys"

func initAppCVC(t *testing.T) (*App, generator.Associated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// fetch the CVC address system variable
	cvcAddr := address.Address{}
	err := app.System(sv.CommandValidatorChangeAddressName, &cvcAddr)
	require.NoError(t, err)
	assc[cvcKeys], err = MockSystemAccount(app, cvcAddr)
	require.NoError(t, err)

	// we require that a node be previously registered
	modify(t, nodeAddress.String(), app, func(acct *backing.AccountData) {
		acct.Balance = 1000 * constants.NapuPerNdau
		acct.ValidationKeys = []signature.PublicKey{nodePublic}
	})
	err = app.UpdateStateImmediately(app.registerNode(nodeAddress, nil, nodePublic))
	require.NoError(t, err)

	// ensure node is primary staker to node rules account
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.Balance = 1000 * constants.NapuPerNdau
	})
	require.NoError(t, err)

	return app, assc
}

func TestCVCIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cvc := NewCommandValidatorChange(
				nodeAddress,
				1,
				1,
				private,
			)

			cvcBytes, err := metatx.Marshal(cvc, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: cvcBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestCVCIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppCVC(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cvc := NewCommandValidatorChange(
		nodeAddress,
		1,
		1,
		private,
	)

	cvcBytes, err := metatx.Marshal(cvc, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: cvcBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCVCIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.CommandValidatorChangeAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 cvc keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cvc := NewCommandValidatorChange(
				nodeAddress,
				1,
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, cvc)

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

// we can't really test that CVC transactions actually update the validator
// set within a unit test: the actual validator set, the one which matters,
// is stored in tendermint. To actually test this behavior, we need to write
// integration tests.
//
// Here's what we _can_ do: we can write tests which ensure that when a
// CVC transactable is sent up, the metanode does the right thing, and
// sends the right update list along with the relevant EndBlock response.
//
// Those tests live below, along with some test helpers to make them work.

func toVals(t *testing.T, app *App, cvcs []CommandValidatorChange) []abci.ValidatorUpdate {
	state := app.GetState().(*backing.State)
	vus := make([]abci.ValidatorUpdate, 0, len(cvcs))
	for _, cvc := range cvcs {
		vup, err := cvc.ToValidator(state)
		require.NoError(t, err)
		require.NotNil(t, vup)
		vus = append(vus, *vup)
	}
	return vus
}

// send every update in the list of validator changes to the metanode,
// and ensure that the metanode has kept track of it and returns it in
// the EndBlock transaction
func updateValidators(t *testing.T, app *App, updates ...CommandValidatorChange) {
	metatxs := make([]metatx.Transactable, len(updates))
	for i := 0; i < len(updates); i++ {
		metatxs[i] = metatx.Transactable(&updates[i])
	}

	resps, ebResp := deliverTxsContext(t, app, metatxs, ddc(t))
	for _, resp := range resps {
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}

	state := app.GetState().(*backing.State)
	actual := ebResp.GetValidatorUpdates()
	expect := make([]abci.ValidatorUpdate, len(updates))
	for i := 0; i < len(updates); i++ {
		vup, err := updates[i].ToValidator(state)
		require.NoError(t, err)
		require.NotNil(t, vup)
		expect[i] = *vup
	}

	t.Logf("expect: %q", expect)
	t.Logf("actual: %q", actual)
	require.ElementsMatch(t, expect, ebResp.GetValidatorUpdates())

	app.Commit()
}

func initAppCVCValidators(
	t *testing.T,
	valQty int,
) (app *App, ma generator.Associated, vcs []CommandValidatorChange) {
	app, ma = initAppCVC(t)
	state := app.GetState().(*backing.State)

	vcs = make([]CommandValidatorChange, 0, valQty)
	validators := make([]abci.ValidatorUpdate, 0, valQty)

	for i := 0; i < valQty; i++ {
		// generate a node
		pub, pvt, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		addr, err := address.Generate(address.KindUser, pub.KeyBytes())
		require.NoError(t, err)
		// we require that a node be previously registered
		modify(t, addr.String(), app, func(acct *backing.AccountData) {
			acct.Balance = 1000 * constants.NapuPerNdau
			acct.ValidationKeys = []signature.PublicKey{pub}
		})
		err = app.UpdateStateImmediately(app.registerNode(addr, nil, pub))
		require.NoError(t, err)

		cvc := NewCommandValidatorChange(
			addr,
			1+int64(i),
			uint64(i)+1,
			pvt,
		)
		vcs = append(vcs, *cvc)
		vup, err := cvc.ToValidator(state)
		require.NoError(t, err)
		require.NotNil(t, vup)
		validators = append(validators, *vup)
	}

	// set up these validators as the initial ones on the chain
	app.InitChain(abci.RequestInitChain{Validators: validators})

	return
}

func TestCommandValidatorChangeInit(t *testing.T) {
	initAppCVCValidators(t, 1)
}

func TestCommandValidatorChangeInitChain(t *testing.T) {
	qtyVals := 10
	app, _, cvcs := initAppCVCValidators(t, qtyVals)

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(t, app, cvcs)),
		actualValidators,
	)
}

func TestCommandValidatorChangeAddValidator(t *testing.T) {
	const qtyVals = 1
	app, ma, cvcs := initAppCVCValidators(t, qtyVals)

	// add a validator
	newCVC := NewCommandValidatorChange(
		nodeAddress,
		1,
		qtyVals+1,
		ma[cvcKeys].([]signature.PrivateKey)...,
	)
	require.NotNil(t, newCVC)
	cvcs = append(cvcs, *newCVC)
	updateValidators(t, app, *newCVC)

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(t, app, cvcs)),
		actualValidators,
	)
}

func TestCommandValidatorChangeRemoveValidator(t *testing.T) {
	const qtyVals = 2
	app, ma, cvcs := initAppCVCValidators(t, qtyVals)

	// remove a validator
	cvc := cvcs[0]
	cvc.Power = 0
	cvc.Signatures = make([]signature.Signature, 0, 1)
	cvc.Sequence = qtyVals + 1

	cvcKeys := ma[cvcKeys].([]signature.PrivateKey)
	cvc.Signatures = []signature.Signature{cvcKeys[0].Sign(cvc.SignableBytes())}

	cvcs = cvcs[1:]
	updateValidators(t, app, cvc)

	actualValidators, err := app.Validators()
	require.NoError(t, err)
	metast.ValidatorsAreEquivalent(
		t,
		metast.ValUpdatesToVals(t, toVals(t, app, cvcs)),
		actualValidators,
	)
}

func TestValidCVCIsValid(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	cvc := NewCommandValidatorChange(
		nodeAddress,
		1,
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, cvc)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestCVCNodeMustBeActive(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	// deactivate the node
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		node := state.Nodes[nodeAddress.String()]
		node.Active = false
		state.Nodes[nodeAddress.String()] = node
		return state, nil
	})

	cvc := NewCommandValidatorChange(
		nodeAddress,
		1,
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, cvc)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCVCIsInvalidWhenMaxValidatorsSet(t *testing.T) {
	app, assc := initAppCVC(t)
	privateKeys := assc[cvcKeys].([]signature.PrivateKey)

	cvc := NewCommandValidatorChange(
		nodeAddress,
		1,
		1,
		privateKeys...,
	)

	// ensure the tx _would_ be valid
	cvcBytes, err := metatx.Marshal(cvc, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: cvcBytes})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// set the sysvar
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		var err error
		state.Sysvars[sv.NodeMaxValidators], err = wkt.Uint64(100).MarshalMsg(nil)
		require.NoError(t, err)
		return state, nil
	})

	// ensure the tx is not valid
	resp = app.CheckTx(abci.RequestCheckTx{Tx: cvcBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

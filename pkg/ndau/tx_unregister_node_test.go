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
	metast "github.com/ndau/metanode/pkg/meta/state"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

func initAppUnregisterNode(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	ensureRecent(t, app, targetAddress.String())
	// ensure the target address is self-staked at the beginning of the test
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
		acct.Balance = 1000 * constants.NapuPerNdau
	})

	// ensure the target address is in the node list
	app.UpdateState(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)

		if state.Nodes == nil {
			state.Nodes = make(map[string]backing.Node)
		}

		state.Nodes[targetAddress.String()] = backing.Node{
			Active: true,
			Key:    targetPublic,
		}

		return state, nil
	})

	// add a costaker: transferAddress
	//	err := app.Stake(1, targetAddress, transferAddress, nra, nil)
	//	require.NoError(t, err)

	return app
}

func TestUnregisterNodeAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	// the address is invalid, but NewUnregisterNode doesn't validate this
	rn := NewUnregisterNode(addr, 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	app := initAppUnregisterNode(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	rn = NewUnregisterNode(fakeTarget, 1, transferPrivate)
	ctkBytes, err = tx.Marshal(rn, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidUnregisterNode(t *testing.T) {
	app := initAppUnregisterNode(t)

	rn := NewUnregisterNode(targetAddress, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestUnregisterNodeMustBeANode(t *testing.T) {
	app := initAppUnregisterNode(t)

	// targetAddress points to a node; transferAddress does not
	rn := NewUnregisterNode(transferAddress, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnregisterNodeDeductsTxFee(t *testing.T) {
	app := initAppUnregisterNode(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		rn := NewUnregisterNode(
			targetAddress,
			uint64(i)+1,
			transferPrivate,
		)

		resp := deliverTxWithTxFee(t, app, rn)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestUnregisterNodeRemovesValidatorPower(t *testing.T) {
	app := initAppUnregisterNode(t)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		var err error
		state.Sysvars[sv.NodeMaxValidators], err = wkt.Uint64(10).MarshalMsg(nil)
		require.NoError(t, err)
		return state, nil
	})

	rn := NewUnregisterNode(targetAddress, 1, transferPrivate)
	resp, reb := deliverTxContext(t, app, rn, ddc(t))
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.NotEmpty(t, reb.ValidatorUpdates)

	expectVU := abci.Ed25519ValidatorUpdate(targetPublic.KeyBytes(), 0)
	require.Equal(t, reb.ValidatorUpdates[0], expectVU)
}

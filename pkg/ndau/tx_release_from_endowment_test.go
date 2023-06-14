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
	metast "github.com/ndau/metanode/pkg/meta/state"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	generator "github.com/ndau/system_vars/pkg/genesis.generator"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/oneiro-ndev/tendermint.0.32.3/abci/types"
)

const (
	rfeKeys    = "rfe private keys"
	sysvarKeys = "sysvar private keys"
)

func initAppRFE(t *testing.T) (*App, generator.Associated) {
	return initAppRFEWithIndex(t, "", -1)
}

func initAppRFEWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	*App, generator.Associated,
) {
	app, assc := initAppWithIndex(t, indexAddr, indexVersion)
	app.InitChain(abci.RequestInitChain{})

	// Fetch the sysvar address system variable.
	sysvarAddr := address.Address{}
	err := app.System(sv.SetSysvarAddressName, &sysvarAddr)
	require.NoError(t, err)
	assc[sysvarKeys], err = MockSystemAccount(app, sysvarAddr)
	require.NoError(t, err)

	// fetch the RFE address system variable
	rfeAddr := address.Address{}
	err = app.System(sv.ReleaseFromEndowmentAddressName, &rfeAddr)
	require.NoError(t, err)
	assc[rfeKeys], err = MockSystemAccount(app, rfeAddr)
	require.NoError(t, err)

	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.TotalRFE = 100 * constants.NapuPerNdau
		return state, nil
	})

	return app, assc
}

func TestRFEIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rfe := NewReleaseFromEndowment(
				targetAddress,
				math.Ndau(1),
				1,
				private,
			)

			rfeBytes, err := tx.Marshal(rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestRFEIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRFE(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	rfe := NewReleaseFromEndowment(
		targetAddress,
		math.Ndau(1),
		1,
		private,
	)

	rfeBytes, err := tx.Marshal(rfe, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRFEAddsNdauToExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				ad.Balance = math.Ndau(10)
			})

			rfe := NewReleaseFromEndowment(
				targetAddress,
				math.Ndau(1),
				uint64(i+1),
				private,
			)

			rfeBytes, err := tx.Marshal(rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
				Time: time.Now(),
			}})
			dresp := app.DeliverTx(abci.RequestDeliverTx{Tx: rfeBytes})
			app.EndBlock(abci.RequestEndBlock{})
			app.Commit()

			require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				require.Equal(t, math.Ndau(11), ad.Balance)
			})
		})
	}
}

func TestValidRFEAddsNdauToNonExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)

			targetAddress, err := address.Generate(address.KindUser, public.KeyBytes())
			require.NoError(t, err)

			rfe := NewReleaseFromEndowment(
				targetAddress,
				math.Ndau(1),
				uint64(i+1),
				private,
			)

			rfeBytes, err := tx.Marshal(rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: rfeBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
				Time: time.Now(),
			}})
			dresp := app.DeliverTx(abci.RequestDeliverTx{Tx: rfeBytes})
			app.EndBlock(abci.RequestEndBlock{})
			app.Commit()

			require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				require.Equal(t, math.Ndau(1), ad.Balance)
			})
		})
	}
}

func TestRFEIsValidOnlyWithSufficientTxFee(t *testing.T) {
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
			rfe := NewReleaseFromEndowment(
				targetAddress,
				math.Ndau(1),
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, rfe)

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

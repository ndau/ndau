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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func getSSVAddress(t *testing.T, app *App) address.Address {
	ssvA := address.Address{}
	err := app.System(sv.SetSysvarAddressName, &ssvA)
	require.NoError(t, err)
	err = ssvA.Revalidate()
	require.NoError(t, err)
	return ssvA
}

func initAppSetSysvar(t *testing.T) (app *App, pvts []signature.PrivateKey) {
	app, _ = initApp(t)

	app.InitChain(abci.RequestInitChain{})

	pvts, err := MockSystemAccount(app, getSSVAddress(t, app))
	require.NoError(t, err)

	return
}

func TestValidSetSysvar(t *testing.T) {
	app, privateKeys := initAppSetSysvar(t)

	ssv := NewSetSysvar("foo", []byte("bar"), 1, privateKeys...)

	resp := deliverTx(t, app, ssv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestSetSysvarIsValidWithSingleKey(t *testing.T) {
	app, privateKeys := initAppSetSysvar(t)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ssv := NewSetSysvar("foo", []byte("bar"), 1+uint64(i), private)
			resp := deliverTx(t, app, ssv)
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestSetSysvarKeyMustValidate(t *testing.T) {
	app, privateKeys := initAppSetSysvar(t)

	for i := 0; i < len(privateKeys); i++ {
		kb := make([]byte, len(privateKeys[i].KeyBytes()))
		copy(kb, privateKeys[i].KeyBytes())

		idx := rand.Intn(len(kb))
		bidx := rand.Intn(8)

		kb[idx] ^= 1 << uint(bidx)

		private, err := signature.RawPrivateKey(privateKeys[i].Algorithm(), kb, nil)
		require.NoError(t, err)

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ssv := NewSetSysvar("foo", []byte("bar"), 1+uint64(i), *private)
			resp := deliverTx(t, app, ssv)
			require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
		})
	}
}

func TestValidSetSysvarChaincode(t *testing.T) {
	app, privateKeys := initAppSetSysvar(t)

	cc := vm.MiniAsm("handler 0 zero enddef")
	cb := wkt.Bytes(cc.Bytes())
	scriptData, err := cb.MarshalMsg(nil)
	require.NoError(t, err)

	t.Run("constructed", func(t *testing.T) {
		t.Logf("encoded: %x", cb)
		t.Logf("base64:  %s", base64.StdEncoding.EncodeToString(cb))
		ad, _ := app.getAccount(getSSVAddress(t, app))
		ssv := NewSetSysvar(sv.SIBScriptName, scriptData, ad.Sequence+1, privateKeys...)
		resp := deliverTx(t, app, ssv)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("manual msgp", func(t *testing.T) {
		data, err := hex.DecodeString("c404a0002088")
		require.NoError(t, err)

		t.Logf("expected:          %x", scriptData)
		t.Logf("externally msgp'd: %x", data)
		require.Equal(t, scriptData, data)
		ad, _ := app.getAccount(getSSVAddress(t, app))
		ssv := NewSetSysvar(sv.SIBScriptName, data, ad.Sequence+1, privateKeys...)
		resp := deliverTx(t, app, ssv)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("real world", func(t *testing.T) {
		// Turns out that the real-world case was mis-encoding the script
		data, err := base64.StdEncoding.DecodeString("xASgACCI")
		require.NoError(t, err)
		t.Logf("constructed: %s", base64.StdEncoding.EncodeToString(scriptData))
		t.Logf("real world:  %s", base64.StdEncoding.EncodeToString(data))
		require.Equal(t, scriptData, data)

		ad, _ := app.getAccount(getSSVAddress(t, app))
		ssv := NewSetSysvar("SIBScript", data, ad.Sequence+1, privateKeys...)
		resp := deliverTx(t, app, ssv)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
}

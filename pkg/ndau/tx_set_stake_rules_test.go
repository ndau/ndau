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

	"github.com/ndau/chaincode/pkg/vm"
	"github.com/ndau/metanode/pkg/meta/app/code"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppSetStakeRules(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	ensureRecent(t, app, targetAddress.String())
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
	})

	return app
}

func TestSetStakeRulesAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	// the address is invalid, but NewSetStakeRules doesn't validate this
	cv := NewSetStakeRules(addr, []byte{}, 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	app := initAppSetStakeRules(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	cv = NewSetStakeRules(fakeTarget, []byte{}, 1, transferPrivate)
	ctkBytes, err = tx.Marshal(cv, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidSetStakeRules(t *testing.T) {
	app := initAppSetStakeRules(t)

	cv := NewSetStakeRules(targetAddress, []byte{}, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	t.Run("empty stake rules must be nil", func(t *testing.T) {
		ad, exists := app.getAccount(targetAddress)
		require.True(t, exists)
		require.Nil(t, ad.StakeRules, "empty stake rules must be nil")
	})
}

func TestSetStakeRulesValidatesChaincode(t *testing.T) {
	app := initAppSetStakeRules(t)

	type testcase struct {
		rules []byte
		valid bool
	}
	cases := []testcase{
		{[]byte{0xde, 0xad, 0xbe, 0xef}, false},
		{vm.MiniAsm("handler 0 zero enddef").Bytes(), true},
	}

	for _, tt := range cases {
		negation := ""
		if !tt.valid {
			negation = "not "
		}
		t.Run(fmt.Sprintf("expect %x is %schaincode", tt.rules, negation), func(t *testing.T) {
			cv := NewSetStakeRules(targetAddress, tt.rules, 1, transferPrivate)
			resp := deliverTx(t, app, cv)
			if tt.valid {
				require.Equal(t, code.OK, code.ReturnCode(resp.Code))
			} else {
				require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
			}

			if tt.valid {
				ad, exists := app.getAccount(targetAddress)
				require.True(t, exists)
				require.Equal(t, tt.rules, ad.StakeRules.Script)
			}
		})
	}
}

func TestSetStakeRulesDeductsTxFee(t *testing.T) {
	app := initAppSetStakeRules(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		cv := NewSetStakeRules(
			targetAddress,
			[]byte{},
			uint64(i)+1,
			transferPrivate,
		)

		resp := deliverTxWithTxFee(t, app, cv)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestSetStakeRulesInvalidWhileOthersUseExistingRules(t *testing.T) {
	app := initAppSetStakeRules(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1 * constants.NapuPerNdau
	})
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.StakeRules = &backing.StakeRules{
			Script:  []byte{0xab, 0xcd},
			Inbound: make(map[string]uint64),
		}
	})
	err := app.UpdateStateImmediately(app.Stake(
		1*constants.NapuPerNdau, sourceAddress,
		targetAddress, targetAddress, nil))
	require.NoError(t, err)

	cv := NewSetStakeRules(targetAddress, []byte{}, 1, transferPrivate)
	resp := deliverTx(t, app, cv)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	require.Contains(t, resp.Log, "cannot change stake rules")
}

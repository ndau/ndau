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
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

// initAppUnstake sets up initial conditions:
// rules account is a proper rules account
// node account is costaked with source account to rules account
// source account is primary staked to rules account

func TestValidResolveStakeTxIsValid(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)
	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)

	// tx must be valid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestResolveStakeTargetValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)
	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)

	// make the account field invalid
	tx.Target = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestResolveStakeStakeRulesValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)
	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)

	// make the account field invalid
	tx.Rules = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestResolveStakeSequenceValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)
	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 0, private)

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestResolveStakeSignatureValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)
	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)

	// flip a single bit in the signature
	sigBytes := tx.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	tx.Signatures[0] = *wrongSignature

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestResolveStakeDeductsTxFee(t *testing.T) {
	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			app, assc, rulesAcct := initAppUnstake(t)
			private := assc[rulesPrivate].(signature.PrivateKey)

			modify(t, rulesAcct.String(), app, func(ad *backing.AccountData) {
				ad.Balance = math.Ndau(i)
			})

			tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1+uint64(i), private)

			resp := deliverTxWithTxFee(t, app, tx)

			var expect code.ReturnCode
			if i == 0 {
				expect = code.InvalidTransaction
			} else {
				expect = code.OK
			}
			require.Equal(t, expect, code.ReturnCode(resp.Code))
		})
	}
}

func TestResolveStakeOfCostakeIsInvalid(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)

	tx := NewResolveStake(nodeAddress, rulesAcct, 0, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestResolveStakeOfPrimaryStakeIsValid(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)

	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestResolveStakeChangesAppState(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[rulesPrivate].(signature.PrivateKey)

	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)

	// source is a primary staker; this stake must have been resolved
	sourceData := state.Accounts[source]
	if sourceData.Holds != nil {
		require.Empty(t, sourceData.Holds)
	}

	// node is a costaker; this stake must have been resolved
	nodeData := state.Accounts[nodeAddress.String()]
	if nodeData.Holds != nil {
		require.Empty(t, nodeData.Holds)
	}

	// must have updated inbound stake list
	rulesData := state.Accounts[rulesAcct.String()]
	require.NotZero(t, rulesData) // must exist
	require.NotNil(t, rulesData.StakeRules)
	require.Empty(t, rulesData.StakeRules.Inbound)
}

func TestResolveTransfersPayment(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	// modify the rules acct chaincode to return one napu payment due
	modify(t, rulesAcct.String(), app, func(ad *backing.AccountData) {
		ad.StakeRules.Script = vm.MiniAsm("handler 0 one enddef").Bytes()
	})
	private := assc[rulesPrivate].(signature.PrivateKey)

	state := app.GetState().(*backing.State)
	oldsrc := state.Accounts[source]
	oldrules := state.Accounts[rulesAcct.String()]

	tx := NewResolveStake(sourceAddress, rulesAcct, 0, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state = app.GetState().(*backing.State)

	// source is a primary staker; this stake must have been resolved
	sourceData := state.Accounts[source]
	require.Equal(t, oldsrc.Balance-1, sourceData.Balance, "source must pay 1")

	// must have updated inbound stake list
	rulesData := state.Accounts[rulesAcct.String()]
	require.Equal(t, oldrules.Balance+1, rulesData.Balance, "rules must receive 1")
}

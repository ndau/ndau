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
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppDelegate(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[eaiNode] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	return app, private
}

func TestValidDelegateTxIsValid(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	// d must be valid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestDelegateAccountValidates(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	// make the account field invalid
	d.Target = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateDelegateValidates(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	// make the account field invalid
	d.Node = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateSequenceValidates(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 0, private)

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateSignatureValidates(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	// flip a single bit in the signature
	sigBytes := d.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	d.Signatures[0] = *wrongSignature

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateChangesAppState(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's delegation node
	require.Equal(t, &nodeAddress, state.Accounts[source].DelegationNode)

	// we must have added the source to the node's delegation responsibilities
	require.Contains(t, state.Delegates, eaiNode)
	require.Contains(t, state.Delegates[eaiNode], source)
}

func TestDelegateRemovesPreviousDelegation(t *testing.T) {
	app, private := initAppDelegate(t)
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// now create a new delegation transaction
	// (ensure the new delegate is also an active node)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[dest] = backing.Node{
			Active: true,
		}
		return st, nil
	})
	d = NewDelegate(sourceAddress, destAddress, 2, private)
	resp = deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's delegation node
	require.Equal(t, &destAddress, state.Accounts[source].DelegationNode)

	// we must have added the source to dest's delegation responsibilities
	require.Contains(t, state.Delegates, dest)
	require.Contains(t, state.Delegates[dest], source)

	// we must have removed the source from eaiNode
	require.Contains(t, state.Delegates, eaiNode)
	require.NotContains(t, state.Delegates[eaiNode], source)
}

func TestDelegateDeductsTxFee(t *testing.T) {
	app, private := initAppDelegate(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewDelegate(sourceAddress, nodeAddress, 1+uint64(i), private)

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

func TestDelegateNodeMustBeActive(t *testing.T) {
	app, private := initAppDelegate(t)
	// ensure the node isn't active
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[eaiNode] = backing.Node{
			Active: false,
		}
		return st, nil
	})

	d := NewDelegate(sourceAddress, nodeAddress, 1, private)

	// d must be valid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

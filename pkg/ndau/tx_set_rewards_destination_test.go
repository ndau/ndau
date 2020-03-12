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
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestValidSetRewardsDestinationTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// srt must be valid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestSetRewardsDestinationAccountValidates(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// make the account field invalid
	srt.Target = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationDestinationValidates(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// make the account field invalid
	srt.Destination = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 0, private)

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// flip a single bit in the signature
	sigBytes := srt.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	srt.Signatures[0] = *wrongSignature

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationChangesAppState(t *testing.T) {
	app, private := initAppTx(t)
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	resp := deliverTx(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Equal(t, &destAddress, state.Accounts[source].RewardsTarget)
	// we must have updated the dest's inbound rewards targets
	require.Equal(t, []address.Address{sourceAddress}, state.Accounts[dest].IncomingRewardsFrom)

	// resetting to source address saves as "nil" dest address
	srt = NewSetRewardsDestination(sourceAddress, sourceAddress, 2, private)
	resp = deliverTx(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	state = app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Nil(t, state.Accounts[source].RewardsTarget)
	// we mut have removed the source from the dest's inbound rewards targets
	require.Empty(t, state.Accounts[dest].IncomingRewardsFrom)
}

func TestSetRewardsDestinationInvalidIfDestinationAlsoSends(t *testing.T) {
	app, private := initAppTx(t)

	// when the destination has a rewards target set...
	modify(t, dest, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &nodeAddress
	})

	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationInvalidIfSourceAlsoReceives(t *testing.T) {
	app, private := initAppTx(t)

	// when the source is receiving rewards from another account
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.IncomingRewardsFrom = []address.Address{nodeAddress}
	})

	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestReSetRewardsDestinationChangesAppState(t *testing.T) {
	// set up accounts
	app, private := initAppTx(t)
	tA, err := address.Validate(settled)
	require.NoError(t, err)

	// set up fixture: sourceAddress -> nodeAddress
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &nodeAddress
	})
	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.IncomingRewardsFrom = []address.Address{sourceAddress, tA}
	})

	// deliver transaction
	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)
	resp := deliverTx(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Equal(t, &destAddress, state.Accounts[source].RewardsTarget)
	// we must have updated the dest's inbound rewards targets
	require.Equal(t, []address.Address{sourceAddress}, state.Accounts[dest].IncomingRewardsFrom)
	// we must have removed the prev target's inbound targets
	require.Equal(t, []address.Address{tA}, state.Accounts[eaiNode].IncomingRewardsFrom)
}

func TestNotifiedDestinationsAreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	// fixture: destination must be notified
	modify(t, dest, app, func(ad *backing.AccountData) {
		uo := math.Timestamp(app.BlockTime() + 1)
		ad.Lock = backing.NewLock(math.Duration(2), eai.DefaultLockBonusEAI)
		ad.Lock.UnlocksOn = &uo
	})

	srt := NewSetRewardsDestination(sourceAddress, destAddress, 1, private)

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := uint64(0); i < 2; i++ {
		tx := NewSetRewardsDestination(sourceAddress, destAddress, 1+i, private)

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

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
	"time"

	"github.com/ndau/chaincode/pkg/vm"
	"github.com/ndau/metanode/pkg/meta/app/code"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndau/pkg/query"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBurnsWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []math.Ndau{0, -1, -2} {
		tr := NewBurn(sourceAddress, negQty, uint64(idx+1), private)
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestBurnsFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = backing.NewLock(90*math.Day, eai.DefaultLockBonusEAI)
	})

	tr := NewBurn(sourceAddress, 1, 1, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestBurnsFromLockedButExpiredAddressesAreValid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		twoDaysAgo := now.Sub(math.Duration(2 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &twoDaysAgo
	})

	tr := NewBurn(sourceAddress, 1, 1, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestBurnsFromNotifiedAddressesAreInvalid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		tomorrow := now.Add(math.Duration(1 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &tomorrow
	})

	tr := NewBurn(sourceAddress, 1, 1, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestBurnsDeductBalanceFromSource(t *testing.T) {
	app, private := initAppTx(t)

	var initialSourceNdau int64
	modifySource(t, app, func(src *backing.AccountData) {
		initialSourceNdau = int64(src.Balance)
	})

	const deltaNapu = 50 * constants.QuantaPerUnit

	tr := NewBurn(sourceAddress, deltaNapu, 1, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		require.Equal(t, initialSourceNdau-deltaNapu, int64(src.Balance))
	})
}

func TestBurnSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := NewBurn(sourceAddress, 1, 1, private)
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestInvalidBurnDoesntAffectBalance(t *testing.T) {
	app, private := initAppTx(t)
	var (
		beforeSrc math.Ndau
		afterSrc  math.Ndau
	)
	modifySource(t, app, func(src *backing.AccountData) {
		beforeSrc = src.Balance
	})

	// invalid: sequence 0
	tr := NewBurn(sourceAddress, 1, 0, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		afterSrc = src.Balance
	})

	require.Equal(t, beforeSrc, afterSrc)
}

func TestBurnsOfMoreThanSourceBalanceAreInvalid(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(src *backing.AccountData) {
		src.Balance = 1 * constants.QuantaPerUnit
	})
	tr := NewBurn(sourceAddress, 2*constants.QuantaPerUnit, 1, private)
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestBurnSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := NewBurn(sourceAddress, 1, 0, private)
	resp := deliverTx(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := NewBurn(sourceAddress, 1, 1, private)
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestBurnWithExpiredRecoursesWorks(t *testing.T) {
	// setup app
	app, key, ts := initAppRecourse(t)
	require.True(t, app.BlockTime().Compare(ts) >= 0)
	tn := ts.Add(1 * math.Second)

	// generate transfer
	// because the recourse period holds have ended
	// this should succeed
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	tr := NewBurn(
		sourceAddress,
		1,
		1, key,
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestBurnWithUnexpiredRecoursesFails(t *testing.T) {
	// setup app
	app, key, ts := initAppRecourse(t)
	// set app time to a day before the recourse period expiry time
	tn := ts.Add(math.Duration(-24 * 3600 * math.Second))

	// generate transfer
	// because the recourse period holds have ended
	// this should fail
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	tr := NewBurn(
		sourceAddress,
		1,
		1, key,
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidationScriptValidatesBurns(t *testing.T) {
	app, private := initAppTx(t)
	public2, private2, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// this script just ensures that the first validation key
	// is used, no matter how many keys are included
	script := vm.MiniAsm("handler 0 one and not enddef")

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.ValidationScript = script.Bytes()
		ad.ValidationKeys = append(ad.ValidationKeys, public2)
	})

	t.Run("only first key", func(t *testing.T) {
		tr := NewBurn(sourceAddress, 123, 1, private)
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := NewBurn(sourceAddress, 123, 2, private, private2)
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := NewBurn(sourceAddress, 123, 3, private2, private)
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := NewBurn(sourceAddress, 123, 4, private2)
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestBurnDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	for i := uint64(0); i < 2; i++ {
		modify(t, source, app, func(ad *backing.AccountData) {
			ad.Balance = math.Ndau(1 + i)
		})

		tx := NewBurn(sourceAddress, 1, 1+i, private)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		// this is different from the other TestXDeductsTxFee transactions:
		// we had to change the setup to account for the actual amount to be
		// transferred. The logic here is correct.
		if i > 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestAppSurvivesEmptySignatureBurn(t *testing.T) {
	app, _ := initAppTx(t)

	tx := NewBurn(sourceAddress, 1, 1)

	resp := deliverTx(t, app, tx)
	// must be invalid because no signatures
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestBurnedQtyIsTracked(t *testing.T) {
	app, private := initAppTx(t)

	oldTotalBurned := app.GetState().(*backing.State).TotalBurned

	qty := math.Ndau(50) * constants.NapuPerNdau
	tx := NewBurn(sourceAddress, qty, 1, private)

	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	newTotalBurned := app.GetState().(*backing.State).TotalBurned
	require.Equal(t, math.Ndau(oldTotalBurned+qty), newTotalBurned)
}

func TestBurnedQtyDeductedFromTotalNdau(t *testing.T) {
	app, private := initAppTx(t)

	getTotal := func() math.Ndau {
		resp := app.Query(abci.RequestQuery{
			Path: query.SummaryEndpoint,
			Data: nil,
		})
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		summary := new(query.Summary)
		_, err := summary.UnmarshalMsg(resp.Value)
		require.NoError(t, err)
		return summary.TotalNdau
	}

	oldTotal := getTotal()

	qty := math.Ndau(50) * constants.NapuPerNdau
	tx := NewBurn(sourceAddress, qty, 1, private)

	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// normally, Commit is immediately followed by a new BeginBlock which resets
	// the block height, which would cause the summary to update. However, in
	// this case, it doesn't, because we haven't manually begun a block. Therefore,
	// let's just force recalculation of the summary
	lastSummary.BlockHeight = ^uint64(0)

	newTotal := getTotal()
	require.Equal(t, math.Ndau(oldTotal-qty), newTotal)
}

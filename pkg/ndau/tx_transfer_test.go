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
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppTx(t *testing.T) (*App, signature.PrivateKey) {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// generate the validation key so we can transfer from it
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ensureRecent(t, app, source)

	modifySource(t, app, func(acct *backing.AccountData) {
		// initialize the source address with a bunch of ndau
		acct.Balance = math.Ndau(10000 * constants.QuantaPerUnit)
		acct.ValidationKeys = []signature.PublicKey{public}
	})

	return app, private
}

// generate an app with an account with a bunch of transactions with recourse holds
//
// returns that account's private key, and a timestamp after which all holds
// should be valid
//
// It is guaranteed that all recourse holds expire in the interval (timestamp - 1 day : timestamp)
func initAppRecourse(t *testing.T) (*App, signature.PrivateKey, math.Timestamp) {
	app, _ := initAppTx(t)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// generate the validation key so we can transfer from the acct with recourse holds
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	const qtyRecourses = 10

	ensureRecent(t, app, settled)
	modify(t, settled, app, func(acct *backing.AccountData) {
		// initialize the address with a bunch of ndau
		// incoming funds are added to the balance and the holds;
		// it's just that the available balance is reduced by the sum
		// of the uncleared holds
		for i := 1; i < qtyRecourses; i++ {
			acct.Balance += math.Ndau(i * constants.QuantaPerUnit)
			x := ts.Sub(math.Duration(i))
			acct.Holds = append(acct.Holds, backing.Hold{
				Qty:    math.Ndau(i * constants.QuantaPerUnit),
				Expiry: &x,
			})
		}
		acct.ValidationKeys = []signature.PublicKey{public}
	})

	// add 1 second to the timestamp to get past unix time rounding errors
	tn := constants.Epoch.Add(time.Duration(int64(ts)) * time.Microsecond)
	tn = tn.Add(time.Duration(1 * time.Second))

	// update the app's cached timestamp
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: tn,
	}})
	app.EndBlock(abci.RequestEndBlock{})

	return app, private, ts
}

// update the source account
func modifySource(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, source, app, f)
}

// update the dest account
func modifyDest(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, dest, app, f)
}

func generateTransfer(t *testing.T, qty int64, seq uint64, keys []signature.PrivateKey) *Transfer {
	tr := NewTransfer(
		sourceAddress, destAddress,
		math.Ndau(qty*constants.QuantaPerUnit),
		seq, keys...,
	)
	return tr
}

func TestTransfersWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []int64{0, -1, -2} {
		tr := generateTransfer(t, negQty, uint64(idx+1), []signature.PrivateKey{private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestTransfersFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = backing.NewLock(90*math.Day, eai.DefaultLockBonusEAI)
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersFromLockedButExpiredAddressesAreValid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		twoDaysAgo := now.Sub(math.Duration(2 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &twoDaysAgo
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTransfersFromNotifiedAddressesAreInvalid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		tomorrow := now.Add(math.Duration(1 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &tomorrow
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersUpdateDestWAA(t *testing.T) {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)

	modifyDest(t, app, func(acct *backing.AccountData) {
		acct.Balance = 100 * constants.QuantaPerUnit
		acct.LastWAAUpdate = timestamp.Sub(math.Duration(30 * math.Day))
	})

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTxAt(t, app, tr, timestamp)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// because we're doing integer math on times here, we won't get
	// exact results. The results are _deterministic_, just not
	// _exact_. As such, we need to define success in terms of
	// error margins.
	//
	// Given that we're constrained by tendermint limitations to
	// block times at a resolution of 1 second anyway, it makes sense
	// to require that we calculate the correct second.
	const maxEpsilon = int64(1000) * math.Millisecond
	var epsilon int64
	expect := int64(20 * math.Day)
	// not actually modifying the dest here; this is just the
	// fastest way to get access to the account data
	modifyDest(t, app, func(acct *backing.AccountData) {
		epsilon = expect - int64(acct.WeightedAverageAge)
	})
	if epsilon < 0 {
		epsilon = -epsilon
	}
	require.True(
		t, epsilon < maxEpsilon,
		"must be true: epsilon < maxEpsilon",
		"epsilon", epsilon,
		"max epsilon", maxEpsilon,
	)
}

func TestTransfersUpdateDestLastWAAUpdate(t *testing.T) {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// truncate timestamp to the nearest second:
	// we assume that unix time is what tendermint uses,
	// and unix time is truncated to seconds
	timestamp = math.Timestamp(int64(timestamp) - (int64(timestamp) % math.Second))

	app, private := initAppTx(t)

	modifyDest(t, app, func(acct *backing.AccountData) {
		acct.Balance = 100 * constants.QuantaPerUnit
		acct.LastWAAUpdate = timestamp.Sub(math.Duration(30 * math.Day))
	})

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTxAt(t, app, tr, timestamp)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// not actually modifying the dest here; this is just the
	// fastest way to get access to the account data
	modifyDest(t, app, func(acct *backing.AccountData) {
		require.Equal(t, timestamp, acct.LastWAAUpdate)
	})
}

func TestTransfersDeductBalanceFromSource(t *testing.T) {
	app, private := initAppTx(t)

	var initialSourceNdau int64
	modifySource(t, app, func(src *backing.AccountData) {
		initialSourceNdau = int64(src.Balance)
	})

	const deltaNapu = 50 * constants.QuantaPerUnit

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		require.Equal(t, initialSourceNdau-deltaNapu, int64(src.Balance))
	})
}

func TestTransfersAddBalanceToDest(t *testing.T) {
	app, private := initAppTx(t)

	var initialDestNdau int64
	modifyDest(t, app, func(dest *backing.AccountData) {
		initialDestNdau = int64(dest.Balance)
	})

	const deltaNapu = 123 * constants.QuantaPerUnit

	tr := generateTransfer(t, 123, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifyDest(t, app, func(dest *backing.AccountData) {
		require.Equal(t, initialDestNdau+deltaNapu, int64(dest.Balance))
	})
}

func TestTransfersWhoseSrcAndDestAreEqualAreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	qty := int64(1)
	seq := uint64(1)

	// generate a transfer
	// We need to ensure that the application
	// layer rejects deserialized transfers which are invalid.
	tr := generateTransfer(t, qty, seq, []signature.PrivateKey{private})
	tr.Destination = tr.Source
	bytes := tr.SignableBytes()
	tr.Signatures = []signature.Signature{private.Sign(bytes)}

	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransferSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestInvalidTransactionDoesntAffectAnyBalance(t *testing.T) {
	app, private := initAppTx(t)
	var (
		beforeSrc  math.Ndau
		beforeDest math.Ndau
		afterSrc   math.Ndau
		afterDest  math.Ndau
	)
	modifySource(t, app, func(src *backing.AccountData) {
		beforeSrc = src.Balance
	})
	modifyDest(t, app, func(dest *backing.AccountData) {
		beforeDest = dest.Balance
	})

	// invalid: sequence 0
	tr := generateTransfer(t, 1, 0, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		afterSrc = src.Balance
	})
	modifyDest(t, app, func(dest *backing.AccountData) {
		afterDest = dest.Balance
	})

	require.Equal(t, beforeSrc, afterSrc)
	require.Equal(t, beforeDest, afterDest)
}

func TestTransfersOfMoreThanSourceBalanceAreInvalid(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(src *backing.AccountData) {
		src.Balance = 1 * constants.QuantaPerUnit
	})
	tr := generateTransfer(t, 2, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransferSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransfer(t, 1, 0, []signature.PrivateKey{private})
	resp := deliverTx(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransferWithExpiredRecoursesWorks(t *testing.T) {
	// setup app
	app, key, ts := initAppRecourse(t)
	require.True(t, app.BlockTime().Compare(ts) >= 0)
	tn := ts.Add(1 * math.Second)

	// generate transfer
	// because the recourse period holds have ended
	// this should succeed
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	destAddress := destAddress
	require.NoError(t, err)
	tr := NewTransfer(
		sourceAddress, destAddress,
		math.Ndau(1),
		1, key,
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTransferWithUnexpiredRecoursesFails(t *testing.T) {
	// setup app
	app, key, ts := initAppRecourse(t)
	// set app time to a day before the recourse period expiry time
	tn := ts.Add(math.Duration(-24 * 3600 * math.Second))

	// generate transfer
	// because the recourse period holds have ended
	// this should fail
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	destAddress := destAddress
	require.NoError(t, err)
	tr := NewTransfer(
		sourceAddress, destAddress,
		math.Ndau(1),
		1, key,
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidationScriptValidatesTransfers(t *testing.T) {
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
		tr := generateTransfer(t, 123, 1, []signature.PrivateKey{private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 2, []signature.PrivateKey{private, private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 3, []signature.PrivateKey{private2, private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := generateTransfer(t, 123, 4, []signature.PrivateKey{private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestTransferDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	for i := uint64(0); i < 2; i++ {
		modify(t, source, app, func(ad *backing.AccountData) {
			ad.Balance = math.Ndau(1 + i)
		})

		tx := NewTransfer(sourceAddress, destAddress, 1, 1+i, private)

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

func TestTransferToExchangeAddressHasNoRecoursePeriod(t *testing.T) {
	// Make the dest an exchange account
	context := ddc(t).withExchangeAccount(destAddress)
	app, private := initAppTx(t)

	// first, ensure that the source account has non-0 recourse settings
	sourceAcct, _ := app.getAccount(sourceAddress)
	if sourceAcct.RecourseSettings.Period == 0 {
		modifySource(t, app, func(ad *backing.AccountData) {
			ad.RecourseSettings.Period = math.Day
		})
	}

	// refresh and check
	sourceAcct, _ = app.getAccount(sourceAddress)
	require.NotZero(t, sourceAcct.RecourseSettings.Period)

	// ensure there are no existing holds and no balance in the destination account
	destAcct, _ := app.getAccount(destAddress)
	require.Zero(t, destAcct.Balance)
	require.Empty(t, destAcct.Holds)

	// perform the transfer
	tx := NewTransfer(sourceAddress, destAddress, 1*constants.NapuPerNdau, 1, private)
	resp, _ := deliverTxContext(t, app, tx, context)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// the transfer must have arrived in the balance without creating a hold
	destAcct, _ = app.getAccount(destAddress)
	require.NotZero(t, destAcct.Balance)
	require.Empty(t, destAcct.Holds)
}

func TestAppSurvivesEmptySignatureTransfer(t *testing.T) {
	app, _ := initAppTx(t)

	tx := NewTransfer(sourceAddress, destAddress, 1, 1)

	resp := deliverTx(t, app, tx)
	// must be invalid because no signatures
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransferDestFieldValidates(t *testing.T) {
	app, private := initAppTx(t)

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	// unset the destination
	tr.Destination = address.Address{}
	// re-sign because we just changed the signable bytes
	tr.Signatures = []signature.Signature{metatx.Sign(tr, private)}

	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

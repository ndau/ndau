package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestValidSetRewardsDestinationTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// srt must be valid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestSetRewardsDestinationAccountValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// make the account field invalid
	srt.Source = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationDestinationValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// make the account field invalid
	srt.Destination = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 0, private)

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// flip a single bit in the signature
	sigBytes := srt.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	srt.Signatures[0] = *wrongSignature

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationChangesAppState(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsDestination(sA, dA, 1, private)

	resp := deliverTx(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Equal(t, &dA, state.Accounts[source].RewardsTarget)
	// we must have updated the dest's inbound rewards targets
	require.Equal(t, []address.Address{sA}, state.Accounts[dest].IncomingRewardsFrom)

	// resetting to source address saves as "nil" dest address
	srt = NewSetRewardsDestination(sA, sA, 2, private)
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
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)

	// when the destination has a rewards target set...
	modify(t, dest, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &nA
	})

	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationInvalidIfSourceAlsoReceives(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)

	// when the source is receiving rewards from another account
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.IncomingRewardsFrom = []address.Address{nA}
	})

	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestReSetRewardsDestinationChangesAppState(t *testing.T) {
	// set up accounts
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	tA, err := address.Validate(settled)
	require.NoError(t, err)

	// set up fixture: sA -> nA
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &nA
	})
	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.IncomingRewardsFrom = []address.Address{sA, tA}
	})

	// deliver transaction
	srt := NewSetRewardsDestination(sA, dA, 1, private)
	resp := deliverTx(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Equal(t, &dA, state.Accounts[source].RewardsTarget)
	// we must have updated the dest's inbound rewards targets
	require.Equal(t, []address.Address{sA}, state.Accounts[dest].IncomingRewardsFrom)
	// we must have removed the prev target's inbound targets
	require.Equal(t, []address.Address{tA}, state.Accounts[eaiNode].IncomingRewardsFrom)
}

func TestNotifiedDestinationsAreInvalid(t *testing.T) {
	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	app.blockTime = ts
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)

	// fixture: destination must be notified
	modify(t, dest, app, func(ad *backing.AccountData) {
		uo := math.Timestamp(ts + 1)
		ad.Lock = backing.NewLock(math.Duration(2), eai.DefaultLockBonusEAI)
		ad.Lock.UnlocksOn = &uo
	})

	srt := NewSetRewardsDestination(sA, dA, 1, private)

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsDestinationDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := uint64(0); i < 2; i++ {
		tx := NewSetRewardsDestination(sA, dA, 1+i, private)

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

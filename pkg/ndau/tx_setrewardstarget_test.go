package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestValidSetRewardsTargetTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// srt must be valid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestSetRewardsTargetAccountValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// make the account field invalid
	srt.Account = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsTargetDestinationValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// make the account field invalid
	srt.Destination = address.Address{}
	srt.Signatures = []signature.Signature{private.Sign(srt.SignableBytes())}

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsTargetSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 0, []signature.PrivateKey{private})

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsTargetSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

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

func TestSetRewardsTargetChangesAppState(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	resp := deliverTr(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Equal(t, &dA, state.Accounts[source].RewardsTarget)
	// we must have updated the dest's inbound rewards targets
	require.Equal(t, []address.Address{sA}, state.Accounts[dest].IncomingRewardsFrom)

	// resetting to source address saves as "nil" dest address
	srt = NewSetRewardsTarget(sA, sA, 2, []signature.PrivateKey{private})
	resp = deliverTr(t, app, srt)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	state = app.GetState().(*backing.State)
	// we must have updated the source's rewards target
	require.Nil(t, state.Accounts[source].RewardsTarget)
	// we mut have removed the source from the dest's inbound rewards targets
	require.Empty(t, state.Accounts[dest].IncomingRewardsFrom)
}

func TestSetRewardsTargetInvalidIfDestinationAlsoSends(t *testing.T) {
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

	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSetRewardsTargetInvalidIfSourceAlsoReceives(t *testing.T) {
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

	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// ...srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestReSetRewardsTargetChangesAppState(t *testing.T) {
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
	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, srt)
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
		ad.Lock = &backing.Lock{
			NoticePeriod: math.Duration(2),
			UnlocksOn:    &uo,
		}
	})

	srt := NewSetRewardsTarget(sA, dA, 1, []signature.PrivateKey{private})

	// srt must be invalid
	bytes, err := tx.Marshal(srt, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

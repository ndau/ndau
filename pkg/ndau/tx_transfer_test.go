package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/ndaumath/pkg/address"

	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta.app/meta.state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
)

// Private key:    e283e6899f67fe424fc3dd61a79ed3b0860e9925413ccdcbe25422a89e69267088c3d538395e3945e3e6f267974cae362d70acd0389436288bf99422d69c25bb
// Public key:     88c3d538395e3945e3e6f267974cae362d70acd0389436288bf99422d69c25bb
const source = "ndanp2cieaz6w3viwfdxf5dibrt5u8zmdtdep7w3n7yvqsrc"

// Private key:    e88aace3976894b5b40d0dac5cbaf497f0dfe3459c901ae8da813477a1aa1c6b2e89496b55e40021d4814440b3e0cabbe9302abb99b9fe631f3b55c2a913c3bb
// Public key:     2e89496b55e40021d4814440b3e0cabbe9302abb99b9fe631f3b55c2a913c3bb
const dest = "ndam5v8hpv5b79zbxxcepih8d4km4a3j2ev8dpaegexpdest"

func initAppTx(t *testing.T) (*App, signature.PrivateKey) {
	app := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// generate the transfer key so we can transfer from it
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	modifySource(t, app, func(acct *backing.AccountData) {
		// initialize the source address with a bunch of ndau
		acct.Balance = math.Ndau(1000000 * constants.QuantaPerUnit)
		acct.TransferKey = &public
	})

	return app, private
}

func modify(t *testing.T, addr string, app *App, f func(*backing.AccountData)) {
	err := app.UpdateState(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		acct := state.Accounts[addr]

		f(&acct)

		state.Accounts[addr] = acct
		return state, nil
	})

	require.NoError(t, err)
}

// update the source account
func modifySource(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, source, app, f)
}

// update the dest account
func modifyDest(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, dest, app, f)
}

func deliverTr(t *testing.T, app *App, transfer *Transfer) abci.ResponseDeliverTx {
	return deliverTrAt(t, app, transfer, time.Now().Unix())
}

func deliverTrAt(t *testing.T, app *App, transfer *Transfer, time int64) abci.ResponseDeliverTx {
	bytes, err := tx.TransactableToBytes(transfer, TxIDs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time,
	}})
	resp := app.DeliverTx(bytes)
	t.Log(resp.Log)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return resp
}

func generateTransfer(t *testing.T, qty int64, seq uint64, key signature.PrivateKey) *Transfer {
	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	s, err := address.Validate(source)
	require.NoError(t, err)
	d, err := address.Validate(dest)
	require.NoError(t, err)
	tr, err := NewTransfer(
		ts,
		s, d,
		math.Ndau(qty*constants.QuantaPerUnit),
		seq, key,
	)
	require.NoError(t, err)
	return tr
}

func TestTransfersWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []int64{0, -1, -2} {
		tr := generateTransfer(t, negQty, uint64(idx+1), private)
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestTransfersFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = &backing.Lock{
			NoticePeriod: 90 * math.Day,
		}
	})

	tr := generateTransfer(t, 1, 1, private)
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersFromLockedButExpiredAddressesAreValid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		twoDaysAgo := now.Sub(math.Duration(2 * math.Day))
		acct.Lock = &backing.Lock{
			NoticePeriod: math.Duration(1 * math.Day),
			UnlocksOn:    &twoDaysAgo,
		}
	})

	tr := generateTransfer(t, 1, 1, private)
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTransfersFromNotifiedAddressesAreInvalid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		tomorrow := now.Add(math.Duration(1 * math.Day))
		acct.Lock = &backing.Lock{
			NoticePeriod: math.Duration(1 * math.Day),
			UnlocksOn:    &tomorrow,
		}
	})

	tr := generateTransfer(t, 1, 1, private)
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersUpdateDestWAA(t *testing.T) {
	now := time.Now()
	unixTime := now.Unix()
	timestamp, err := math.TimestampFrom(now)
	require.NoError(t, err)

	app, private := initAppTx(t)

	modifyDest(t, app, func(acct *backing.AccountData) {
		acct.Balance = 100 * constants.QuantaPerUnit
		acct.LastWAAUpdate = timestamp.Sub(math.Duration(30 * math.Day))
	})

	tr := generateTransfer(t, 50, 1, private)
	resp := deliverTrAt(t, app, tr, unixTime)
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
	now := time.Now()
	unixTime := now.Unix()
	timestamp, err := math.TimestampFrom(now)
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

	tr := generateTransfer(t, 50, 1, private)
	resp := deliverTrAt(t, app, tr, unixTime)
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

	tr := generateTransfer(t, 50, 1, private)
	resp := deliverTr(t, app, tr)
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

	tr := generateTransfer(t, 123, 1, private)
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifyDest(t, app, func(dest *backing.AccountData) {
		require.Equal(t, initialDestNdau+deltaNapu, int64(dest.Balance))
	})
}

func TestTransfersWhoseSrcAndDestAreEqualAreInvalid(t *testing.T) {
	app, key := initAppTx(t)

	qty := int64(1)
	seq := uint64(1)

	// generate a transfer
	// this is almost a straight copy-paste of generateTransfer,
	// but we use source as dest as well
	//
	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	s, err := address.Validate(source)
	require.NoError(t, err)
	_, err = NewTransfer(
		ts,
		s, s,
		math.Ndau(qty*constants.QuantaPerUnit),
		seq, key,
	)
	require.Error(t, err)

	// We've just proved that this implementation refuses to generate
	// a transfer for which source and dest are identical.
	//
	// However, what if someone builds one from scratch?
	// We need to ensure that the application
	// layer rejects deserialized transfers which are invalid.
	tr := generateTransfer(t, qty, seq, key)
	tr.Destination = tr.Source
	bytes, err := tr.signableBytes()
	require.NoError(t, err)
	tr.Signature, err = key.Sign(bytes).Marshal()
	require.NoError(t, err)

	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 1, 1, private)
	// I'm almost completely certain that this will be an invalid signature
	tr.Signature = []byte("foo bar bat baz")
	resp := deliverTr(t, app, tr)
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
	tr := generateTransfer(t, 1, 0, private)
	resp := deliverTr(t, app, tr)
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
	tr := generateTransfer(t, 2, 1, private)
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransfer(t, 1, 0, private)
	resp := deliverTr(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransfer(t, 1, 1, private)
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

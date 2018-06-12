package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
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
		acct.TransferKey, err = public.Marshal()
		require.NoError(t, err)
	})

	return app, private
}

// update the source account
func modifySource(t *testing.T, app *App, f func(*backing.AccountData)) {
	state := app.GetState().(*backing.State)
	acct, err := state.GetAccount(app.GetDB(), source)
	require.NoError(t, err)

	f(&acct)

	err = state.UpdateAccount(app.GetDB(), source, acct)
	require.NoError(t, err)
}

// update the dest account
func modifyDest(t *testing.T, app *App, f func(*backing.AccountData)) {
	state := app.GetState().(*backing.State)
	acct, err := state.GetAccount(app.GetDB(), dest)
	require.NoError(t, err)

	f(&acct)

	err = state.UpdateAccount(app.GetDB(), dest, acct)
	require.NoError(t, err)
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
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return resp
}

func generateTransfer(t *testing.T, qty int64, seq uint64, key signature.PrivateKey) *Transfer {
	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	tr, err := NewTransfer(
		ts,
		source, dest,
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
	// I think that half of a second accuracy should be fine
	// for our purposes.
	const maxEpsilon = int64(500) * math.Millisecond
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

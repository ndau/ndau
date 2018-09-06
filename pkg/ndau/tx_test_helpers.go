package ndau

import (
	"testing"
	"time"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/cache"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Private key:    e283e6899f67fe424fc3dd61a79ed3b0860e9925413ccdcbe25422a89e69267088c3d538395e3945e3e6f267974cae362d70acd0389436288bf99422d69c25bb
// Public key:     88c3d538395e3945e3e6f267974cae362d70acd0389436288bf99422d69c25bb
const source = "ndanp2cieaz6w3viwfdxf5dibrt5u8zmdtdep7w3n7yvqsrc"

// Private key:    e88aace3976894b5b40d0dac5cbaf497f0dfe3459c901ae8da813477a1aa1c6b2e89496b55e40021d4814440b3e0cabbe9302abb99b9fe631f3b55c2a913c3bb
// Public key:     2e89496b55e40021d4814440b3e0cabbe9302abb99b9fe631f3b55c2a913c3bb
const dest = "ndam5v8hpv5b79zbxxcepih8d4km4a3j2ev8dpaegexpdest"

// Private key:    73a1955a52d6e7e099607c1bcfe4825fd30632be9780c9d70c836d8c5044546a878f08ca7793c560ca16400e08dfa776cebca90a4d9889524eeeec2fb288cc25
// Public key:     878f08ca7793c560ca16400e08dfa776cebca90a4d9889524eeeec2fb288cc25
const settled = "ndap94hhwyik86x2na9m3hjtq4n5v9uj3qm4tfp4xuyescrw"

// Private key:    e8f080d6f39b0942217a55a4e239cc59b6dfbc48bc3d5e0abebc7da0bf055f57d17516973974aced03ca0ebef33b3798719c596b01a065a0de74e999670e1be5
// Public key:     d17516973974aced03ca0ebef33b3798719c596b01a065a0de74e999670e1be5
const eaiNode = "ndamb84tesvp54vhc63257wifr34zfvyffvi9utqrkruneai"

var (
	targetPrivate signature.PrivateKey
	targetPublic  signature.PublicKey
	targetAddress address.Address

	transferPublic  signature.PublicKey
	transferPrivate signature.PrivateKey
)

func init() {
	var err error
	targetPublic, targetPrivate, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}

	targetAddress, err = address.Generate(address.KindUser, targetPublic.Bytes())
	if err != nil {
		panic(err)
	}

	// require that the public and private keys agree
	testdata := []byte("foo bar bat baz")
	sig := targetPrivate.Sign(testdata)
	if !targetPublic.Verify(testdata, sig) {
		panic("target public and private keys do not agree")
	}

	transferPublic, transferPrivate, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}
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

func deliverTr(t *testing.T, app *App, transfer metatx.Transactable) abci.ResponseDeliverTx {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	return deliverTrAt(t, app, transfer, timestamp)
}

func deliverTrAt(
	t *testing.T,
	app *App,
	transactable metatx.Transactable,
	time math.Timestamp,
) abci.ResponseDeliverTx {
	return deliverTrAtWithSV(
		t,
		app,
		transactable,
		time,
		func(*cache.SystemCache) {},
	)
}

func deliverTrWithSV(
	t *testing.T,
	app *App,
	transactable metatx.Transactable,
	svUpdate func(*cache.SystemCache),
) abci.ResponseDeliverTx {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	return deliverTrAtWithSV(t, app, transactable, timestamp, svUpdate)
}

func deliverTrAtWithSV(
	t *testing.T,
	app *App,
	transactable metatx.Transactable,
	time math.Timestamp,
	svUpdate func(*cache.SystemCache),
) abci.ResponseDeliverTx {
	bytes, err := metatx.Marshal(transactable, TxIDs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.AsTime().Unix(),
	}})
	svUpdate(app.systemCache)
	resp := app.DeliverTx(bytes)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return resp
}

// delivers a transaction with a script which unconditionally sets a tx fee of 1 napu
func deliverTrWithTxFee(t *testing.T, app *App, transactable metatx.Transactable) abci.ResponseDeliverTx {
	return deliverTrWithSV(t, app, transactable, func(systemCache *cache.SystemCache) {
		// set the cached tx fee script to unconditionally return 1
		systemCache.Set(
			sv.TxFeeScriptName,
			// script: oAAaiA==
			wkt.Bytes([]byte{0xa0, 0x00, 0x1a, 0x88}),
		)
	})
}

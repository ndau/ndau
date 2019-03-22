package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/sirupsen/logrus"
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

	childAddress          address.Address
	childPublic           signature.PublicKey
	childPrivate          signature.PrivateKey
	childSignature        signature.Signature
	childSettlementPeriod math.Duration

	transferPublic  signature.PublicKey
	transferPrivate signature.PrivateKey
	transferAddress address.Address

	sourceAddress address.Address
	destAddress   address.Address
	nodeAddress   address.Address
)

func init() {
	var err error
	targetPublic, targetPrivate, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}

	targetAddress, err = address.Generate(address.KindUser, targetPublic.KeyBytes())
	if err != nil {
		panic(err)
	}

	// require that the public and private keys agree
	testdata := []byte("foo bar bat baz")
	sig := targetPrivate.Sign(testdata)
	if !targetPublic.Verify(testdata, sig) {
		panic("target public and private keys do not agree")
	}

	childPublic, childPrivate, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}

	childAddress, err = address.Generate(address.KindUser, childPublic.KeyBytes())
	if err != nil {
		panic(err)
	}

	childAddressBytes := []byte(childAddress.String())
	childSignature = childPrivate.Sign(childAddressBytes)
	if !childPublic.Verify(childAddressBytes, childSignature) {
		panic("child public and private keys do not agree")
	}

	// We'll use the default settlement period for child accounts.  Any negative duration will do.
	childSettlementPeriod = -math.Duration(1)

	transferPublic, transferPrivate, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}

	transferAddress, err = address.Generate(address.KindUser, transferPublic.KeyBytes())
	if err != nil {
		panic(err)
	}

	sourceAddress, err = address.Validate(source)
	if err != nil {
		panic(err)
	}

	destAddress, err = address.Validate(dest)
	if err != nil {
		panic(err)
	}

	nodeAddress, err = address.Validate(eaiNode)
	if err != nil {
		panic(err)
	}
}

func initApp(t *testing.T) (app *App, assc generator.Associated) {
	return initAppWithIndex(t, "", -1)
}

func initAppWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	app *App, assc generator.Associated,
) {
	var err error
	app, assc, err = InitMockAppWithIndex(indexAddr, indexVersion)
	require.NoError(t, err)

	// send log output to the test logger
	logger := logrus.StandardLogger()
	logger.Out = testwriter.New(t)
	app.SetLogger(logger)

	return
}

// app.System depends on app.Height() returning a reasonable value.
// Also, to test all system variable features, we need to be able to
// control what that value is.
//
// Unfortunately, by default, app.Height just crashes before the app
// is fully initialized, which happens at the InitChain transaction.
//
// We need to send InitChain regardless, to set the initial system
// variable cache,
// but that doesn't allow us to control the returned height, and we
// definitely don't want to wait around for the chain to run for some
// number of blocks.
//
// We've solved this by making what should be a private method, public.
// All we have to do now is call it.
func initAppAtHeight(t *testing.T, atHeight uint64) (app *App) {
	app, _ = initApp(t)
	// adjust only if required
	if atHeight != 0 {
		app.SetHeight(atHeight)
	}
	app.InitChain(abci.RequestInitChain{})
	return
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

func modifyNode(t *testing.T, addr string, app *App, f func(*backing.Node)) {
	err := app.UpdateState(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		node := state.Nodes[addr]

		f(&node)

		state.Nodes[addr] = node
		return state, nil
	})

	require.NoError(t, err)
}

func deliverTx(t *testing.T, app *App, tx metatx.Transactable) abci.ResponseDeliverTx {
	resp, _ := deliverTxContext(t, app, tx, ddc(t))
	return resp
}

func deliverTxAt(t *testing.T, app *App, tx metatx.Transactable, at math.Timestamp) abci.ResponseDeliverTx {
	resp, _ := deliverTxContext(t, app, tx, ddc(t).at(at))
	return resp
}

// delivers a transaction with a script which unconditionally sets a tx fee of 1 napu
func deliverTxWithTxFee(t *testing.T, app *App, tx metatx.Transactable) abci.ResponseDeliverTx {
	dc := ddc(t).with(func(sysvars map[string][]byte) {
		sysvars[sv.TxFeeScriptName] = vm.MiniAsm("handler 0 one enddef").Bytes()
	})
	resp, _ := deliverTxContext(t, app, tx, dc)
	return resp
}

type deliveryContext struct {
	t          *testing.T
	ts         math.Timestamp
	svUpdaters []func(svs map[string][]byte)
}

// default delivery context
func ddc(t *testing.T) deliveryContext {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	return deliveryContext{
		t:          t,
		ts:         now,
		svUpdaters: nil,
	}
}

// note: we don't take a pointer, so this copies values, doesn't edit
func (dc deliveryContext) at(ts math.Timestamp) deliveryContext {
	dc.ts = ts
	return dc
}

// add an updater to the list of system variable updaters
func (dc deliveryContext) with(updater func(map[string][]byte)) deliveryContext {
	dc.svUpdaters = append(dc.svUpdaters, updater)
	return dc
}

func (dc deliveryContext) withExchangeAccount(addr address.Address) deliveryContext {
	return dc.with(func(sysvars map[string][]byte) {
		accountAttributes := sv.AccountAttributes{}
		aab, ok := sysvars[sv.AccountAttributesName]
		if ok {
			// modify the existing
			// unpack the struct
			_, err := accountAttributes.UnmarshalMsg(aab)
			require.NoError(dc.t, err)
		}

		// set the attribute
		pattrs := accountAttributes[addr.String()]
		pattrs[sv.AccountAttributeExchange] = struct{}{}
		accountAttributes[addr.String()] = pattrs

		// remarshal
		aab, err := accountAttributes.MarshalMsg(nil)
		require.NoError(dc.t, err)
		// reset
		sysvars[sv.AccountAttributesName] = aab
	})
}

func deliverTxContext(
	t *testing.T,
	app *App,
	tx metatx.Transactable,
	dc deliveryContext,
) (abci.ResponseDeliverTx, abci.ResponseEndBlock) {
	resps, reb := deliverTxsContext(t, app, []metatx.Transactable{tx}, dc)
	require.Equal(t, 1, len(resps), "single transaction must produce single response")
	return resps[0], reb
}

func deliverTxsContext(
	t *testing.T,
	app *App,
	txs []metatx.Transactable,
	dc deliveryContext,
) ([]abci.ResponseDeliverTx, abci.ResponseEndBlock) {
	sysvarCache := make(map[string][]byte)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)

		// copy the current sysvars so we can restore them after the test
		for k, v := range state.Sysvars {
			sysvarCache[k] = v
		}

		// run the sysvar updaters
		for _, updater := range dc.svUpdaters {
			updater(state.Sysvars)
		}

		return state, nil
	})

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: dc.ts.AsTime(),
	}})

	resps := make([]abci.ResponseDeliverTx, 0, len(txs))

	for _, transactable := range txs {
		bytes, err := metatx.Marshal(transactable, TxIDs)
		require.NoError(t, err)

		resp := app.DeliverTx(bytes)
		t.Log(code.ReturnCode(resp.Code))
		if resp.Log != "" {
			t.Log(resp.Log)
		}
		resps = append(resps, resp)
	}
	reb := app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// restore state
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.Sysvars = sysvarCache
		return state, nil
	})

	return resps, reb
}

// setExchangeAccount marks the given address as having the exchange account attribute.
func setExchangeAccount(addr address.Address) {
	accountAttributes := sv.AccountAttributes{}

	attributes := make(map[string]struct{})
	accountAttributes[addr.String()] = attributes

	type Attribute struct{}
	var attribute Attribute
	attributes[sv.AccountAttributeExchange] = attribute
}

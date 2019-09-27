package ndau

import (
	"math/rand"
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppRegisterNode(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	ensureRecent(t, app, targetAddress.String())

	// this ensures the target address exists
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.Balance = 1000 * constants.NapuPerNdau
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
	})

	// ensure node is primary staker to node rules account
	noderules, _ := getRulesAccount(t, app)
	err := app.UpdateStateImmediately(app.Stake(
		1000*constants.NapuPerNdau,
		targetAddress, noderules, noderules,
		nil,
	))
	require.NoError(t, err)

	return app
}

func ensureTargetAddressSyncd(t *testing.T) {
	t.Log("target address", targetAddress)
	t.Log("target public", targetPublic)
	addr, err := address.Generate(targetAddress.Kind(), targetPublic.KeyBytes())
	require.NoError(t, err)
	t.Log("generated address", addr)
	require.Equal(t, targetAddress, addr)
}

func TestRegisterNodeAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	// the address is invalid, but NewRegisterNode doesn't validate this
	rn := NewRegisterNode(addr, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	app := initAppRegisterNode(t)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	rn = NewRegisterNode(fakeTarget, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	ctkBytes, err = tx.Marshal(rn, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeInvalidScript(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{}, targetPublic, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeInvalidPubKey(t *testing.T) {
	app := initAppRegisterNode(t)

	// flip one bit of pub key
	pkb := make([]byte, targetPublic.Algorithm().PublicKeySize())
	copy(pkb, targetPublic.KeyBytes())
	byteIdx := rand.Intn(len(pkb))
	flipBit := byte(1 << uint(rand.Intn(8)))
	pkb[byteIdx] = pkb[byteIdx] ^ flipBit
	pubKey, err := signature.RawPublicKey(targetPublic.Algorithm(), pkb, targetPublic.ExtraBytes())
	require.NoError(t, err)

	ensureTargetAddressSyncd(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, *pubKey, 1, transferPrivate)
	resp := deliverTx(t, app, rn)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRegisterNode(t *testing.T) {
	ensureTargetAddressSyncd(t)
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestRegisterNodeMustBeInactive(t *testing.T) {
	app := initAppRegisterNode(t)
	var err error
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		node := st.Nodes[targetAddress.String()]
		node.Active = true
		st.Nodes[targetAddress.String()] = node
		return st, nil
	})

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}
func TestRegisterNodeDeductsTxFee(t *testing.T) {
	app := initAppRegisterNode(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance++
	})

	for i := 0; i < 2; i++ {
		rn := NewRegisterNode(
			targetAddress,
			[]byte{0xa0, 0x00, 0x88}, targetPublic,
			uint64(i)+1,
			transferPrivate,
		)

		resp := deliverTxWithTxFee(t, app, rn)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestRegisterNodeTargetMustBePrimaryStakerToRulesAccount(t *testing.T) {
	app := initAppRegisterNode(t)
	rulesAcct, _ := getRulesAccount(t, app)
	// unstake the target to set up proper conditions for this test
	err := app.UpdateStateImmediately(app.Unstake(
		1000*constants.NapuPerNdau,
		targetAddress, rulesAcct, rulesAcct,
		0,
	))
	require.NoError(t, err)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRulesAccountMustApproveRegisterNode(t *testing.T) {
	app := initAppRegisterNode(t)
	rulesAcct, _ := getRulesAccount(t, app)
	modify(t, rulesAcct.String(), app, func(ad *backing.AccountData) {
		ad.StakeRules.Script = vm.MiniAsm("handler 0 fail enddef").Bytes()
	})
	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: ctkBytes})
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeMustBeEd25519(t *testing.T) {
	// all the register node initialization assumes we're registering the
	// target node, which isn't appropriate in this case. We have to do it by hand.
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	pubkey, pvtkey, err := signature.Generate(signature.Secp256k1, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindNdau, pubkey.KeyBytes())
	require.NoError(t, err)

	ensureRecent(t, app, addr.String())
	modify(t, addr.String(), app, func(acct *backing.AccountData) {
		acct.Balance = 1000 * constants.NapuPerNdau
		acct.ValidationKeys = []signature.PublicKey{pubkey}
	})

	// ensure node is primary staker to node rules account
	noderules, _ := getRulesAccount(t, app)
	err = app.UpdateStateImmediately(app.Stake(
		1000*constants.NapuPerNdau,
		addr, noderules, noderules,
		nil,
	))
	require.NoError(t, err)

	// now let's create a RegisterNode tx
	rn := NewRegisterNode(addr, []byte{0xa0, 0x00, 0x88}, pubkey, 1, pvtkey)
	resp := deliverTx(t, app, rn)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	require.Contains(t, resp.Log, "ed25519")
}

package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppRegisterNode(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
	})

	return app
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
	rn := NewRegisterNode(addr, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	app := initAppRegisterNode(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	rn = NewRegisterNode(fakeTarget, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)
	ctkBytes, err = tx.Marshal(rn, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeInvalidScript(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{}, "http://1.2.3.4:56789", 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeInvalidRPC(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, "foo bar.baz", 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRegisterNode(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestRegisterNodeMustNotBeStaked(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRegisterNodeStakesSelf(t *testing.T) {
	app := initAppRegisterNode(t)

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)
	resp := deliverTx(t, app, rn)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	_, exists := app.getAccount(targetAddress)
	require.True(t, exists)
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

	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789", 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}
func TestRegisterNodeDeductsTxFee(t *testing.T) {
	app := initAppRegisterNode(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		rn := NewRegisterNode(
			targetAddress,
			[]byte{0xa0, 0x00, 0x88}, "http://1.2.3.4:56789",
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

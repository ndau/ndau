package ndau

import (
	"testing"

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

func initAppUnregisterNode(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// ensure the target address is self-staked at the beginning of the test
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
		acct.Balance = 1000 * constants.NapuPerNdau
	})

	// ensure the target address is in the node list
	app.UpdateState(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)

		if state.Nodes == nil {
			state.Nodes = make(map[string]backing.Node)
		}

		state.Nodes[targetAddress.String()] = backing.Node{
			Active: true,
		}

		return state, nil
	})

	// add a costaker: transferAddress
	//	err := app.Stake(1, targetAddress, transferAddress, nra, nil)
	//	require.NoError(t, err)

	return app
}

func TestUnregisterNodeAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	// the address is invalid, but NewUnregisterNode doesn't validate this
	rn := NewUnregisterNode(addr, 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	app := initAppUnregisterNode(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	rn = NewUnregisterNode(fakeTarget, 1, transferPrivate)
	ctkBytes, err = tx.Marshal(rn, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidUnregisterNode(t *testing.T) {
	app := initAppUnregisterNode(t)

	rn := NewUnregisterNode(targetAddress, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestUnregisterNodeMustBeANode(t *testing.T) {
	app := initAppUnregisterNode(t)

	// targetAddress points to a node; transferAddress does not
	rn := NewUnregisterNode(transferAddress, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(rn, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnregisterNodeDeductsTxFee(t *testing.T) {
	app := initAppUnregisterNode(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		rn := NewUnregisterNode(
			targetAddress,
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

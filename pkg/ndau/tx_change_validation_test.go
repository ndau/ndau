package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	abci "github.com/tendermint/tendermint/abci/types"
)

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

func initAppChangeValidation(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.TransferKeys = []signature.PublicKey{transferPublic}
	})

	return app
}

func TestChangeValidationAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// the address is invalid, but NewChangeValidation doesn't validate this
	cv := NewChangeValidation(addr, []signature.PublicKey{newPublic}, 1, []signature.PrivateKey{transferPrivate})

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(&cv, TxIDs)
	require.NoError(t, err)

	app := initAppChangeValidation(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	cv = NewChangeValidation(fakeTarget, []signature.PublicKey{newPublic}, 1, []signature.PrivateKey{transferPrivate})
	ctkBytes, err = tx.Marshal(&cv, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidChangeValidation(t *testing.T) {
	app := initAppChangeValidation(t)

	// now change the transfer key using the previous transfer key
	newPub, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPub}, 1, []signature.PrivateKey{transferPrivate})
	ctkBytes, err := tx.Marshal(&cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestChangeValidationNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{targetPublic}, 1, []signature.PrivateKey{transferPrivate})
	ctkBytes, err := tx.Marshal(&cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidChangeValidationUpdatesTransferKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, 1, []signature.PrivateKey{transferPrivate})
	resp := deliverTr(t, app, &cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.Bytes(), ad.TransferKeys[0].Bytes())
	})
}

func TestChangeValidationChain(t *testing.T) {
	newPublic, newPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, 1, []signature.PrivateKey{transferPrivate})
	resp := deliverTr(t, app, &cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	cv = NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, 2, []signature.PrivateKey{transferPrivate})
	resp = deliverTr(t, app, &cv)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	newPublic2, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	cv = NewChangeValidation(targetAddress, []signature.PublicKey{newPublic2}, 3, []signature.PrivateKey{newPrivate})
	resp = deliverTr(t, app, &cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

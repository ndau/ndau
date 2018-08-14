package ndau

import (
	"testing"
	"time"

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
}

func initAppCTK(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {})

	return app
}

func TestCTKAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// the address is invalid, but newCTK doesn't validate this
	ctk := NewChangeTransferKeys(addr, newPublic, 1, SigningKeyOwnership, targetPublic, targetPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	app := initAppCTK(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	ctk = NewChangeTransferKeys(fakeTarget, newPublic, 1, SigningKeyOwnership, targetPublic, targetPrivate)
	ctkBytes, err = tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidCTKOwnership(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ctk := NewChangeTransferKeys(targetAddress, newPublic, 1, SigningKeyOwnership, targetPublic, targetPrivate)
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	app := initAppCTK(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestValidCTKTransfer(t *testing.T) {
	transferPublic, transferPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	app := initAppCTK(t)
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.TransferKey = &transferPublic
	})

	// now change the transfer key using the previous transfer key
	newPub, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ctk := NewChangeTransferKeys(targetAddress, newPub, 1, SigningKeyTransfer, transferPublic, transferPrivate)
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestCTKNewTransferKeyNotEqualExistingTransferKey(t *testing.T) {
	transferPublic, transferPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	app := initAppCTK(t)
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.TransferKey = &transferPublic
	})

	ctk := NewChangeTransferKeys(targetAddress, transferPublic, 1, SigningKeyTransfer, transferPublic, transferPrivate)
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCTKNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	transferPublic, transferPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	app := initAppCTK(t)
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.TransferKey = &transferPublic
	})

	ctk := NewChangeTransferKeys(targetAddress, targetPublic, 1, SigningKeyTransfer, transferPublic, transferPrivate)
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidCTKUpdatesTransferKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ctk := NewChangeTransferKeys(targetAddress, newPublic, 1, SigningKeyOwnership, targetPublic, targetPrivate)
	ctkBytes, err := tx.Marshal(&ctk, TxIDs)
	require.NoError(t, err)

	app := initAppCTK(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// apply the transaction
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now().Unix(),
	}})
	dresp := app.DeliverTx(ctkBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	t.Log(dresp.Log)
	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.Bytes(), ad.TransferKey.Bytes())
	})
}

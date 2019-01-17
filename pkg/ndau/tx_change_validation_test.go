package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppChangeValidation(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// this ensures the target address exists
	modify(t, targetAddress.String(), app, func(acct *backing.AccountData) {
		acct.ValidationKeys = []signature.PublicKey{transferPublic}
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
	cv := NewChangeValidation(addr, []signature.PublicKey{newPublic}, []byte{}, 1, transferPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	app := initAppChangeValidation(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	cv = NewChangeValidation(fakeTarget, []signature.PublicKey{newPublic}, []byte{}, 1, transferPrivate)
	ctkBytes, err = tx.Marshal(cv, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidChangeValidation(t *testing.T) {
	app := initAppChangeValidation(t)

	// now change the transfer key using the previous transfer key
	newPub, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPub}, []byte{}, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestChangeValidationNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{targetPublic}, []byte{}, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidChangeValidationUpdatesTransferKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, []byte{}, 1, transferPrivate)
	resp := deliverTx(t, app, cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.KeyBytes(), ad.ValidationKeys[0].KeyBytes())
	})
}

func TestChangeValidationChain(t *testing.T) {
	newPublic, newPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, []byte{}, 1, transferPrivate)
	resp := deliverTx(t, app, cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	cv = NewChangeValidation(targetAddress, []signature.PublicKey{newPublic}, []byte{}, 2, transferPrivate)
	resp = deliverTx(t, app, cv)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	newPublic2, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	cv = NewChangeValidation(targetAddress, []signature.PublicKey{newPublic2}, []byte{}, 3, newPrivate)
	resp = deliverTx(t, app, cv)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestChangeValidationNoValidationKeys(t *testing.T) {
	app := initAppChangeValidation(t)

	cv := NewChangeValidation(targetAddress, []signature.PublicKey{}, []byte{}, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestChangeValidationTooManyValidationKeys(t *testing.T) {
	app := initAppChangeValidation(t)

	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	cv := NewChangeValidation(targetAddress, newKeys, []byte{}, 1, transferPrivate)
	ctkBytes, err := tx.Marshal(cv, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestChangeValidationDeductsTxFee(t *testing.T) {
	app := initAppChangeValidation(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		// now change the transfer key using the previous transfer key
		newPub, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		cv := NewChangeValidation(
			targetAddress,
			[]signature.PublicKey{newPub},
			[]byte{},
			uint64(i)+1,
			transferPrivate,
		)

		resp := deliverTxWithTxFee(t, app, cv)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

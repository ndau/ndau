package ndau

import (
	"fmt"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppClaimAccount(t *testing.T) *App {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	return app
}

func TestClaimAccountAddressFieldValidates(t *testing.T) {
	// flip the bits of the last byte so the address is no longer correct
	addrBytes := []byte(targetAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// ensure that we didn't accidentally create a valid address
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// the address is invalid, but NewSetValidation doesn't validate this
	ca := NewSetValidation(addr, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	ca = NewSetValidation(fakeTarget, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err = tx.Marshal(ca, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimAccount(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestClaimAccountNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app := initAppClaimAccount(t)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{targetPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimAccountUpdatesTransferKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// apply the transaction
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	dresp := app.DeliverTx(ctkBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	t.Log(dresp.Log)
	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.KeyBytes(), ad.ValidationKeys[0].KeyBytes())
	})
}

func TestClaimAccountNoValidationKeys(t *testing.T) {
	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimAccountTooManyValidationKeys(t *testing.T) {
	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	ca := NewSetValidation(targetAddress, targetPublic, newKeys, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimAccountOverwritesOneTransferKey(t *testing.T) {
	app := initAppClaimAccount(t)

	existing, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = []signature.PublicKey{existing}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestClaimAccountCannotOverwriteMoreThanOneTransferKey(t *testing.T) {
	app := initAppClaimAccount(t)

	existing1, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	existing2, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = []signature.PublicKey{existing1, existing2}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	ctkBytes, err := tx.Marshal(ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimAccountDeductsTxFee(t *testing.T) {
	app := initAppClaimAccount(t)
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		newPublic, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		tx := NewSetValidation(
			targetAddress,
			targetPublic,
			[]signature.PublicKey{newPublic},
			[]byte{},
			1+uint64(i),
			targetPrivate,
		)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestClaimAccountDoesntResetWAA(t *testing.T) {
	// inspired by a Real Live Bug!
	app := initAppClaimAccount(t)

	assertExistsAndNonzeroWAAUpdate := func(expectExists bool) {
		resp := app.Query(abci.RequestQuery{
			Path: query.AccountEndpoint,
			Data: []byte(targetAddress.String()),
		})
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		require.Equal(t, fmt.Sprintf(query.AccountInfoFmt, expectExists), resp.Info)

		accountData := new(backing.AccountData)
		_, err := accountData.UnmarshalMsg(resp.Value)
		require.NoError(t, err)

		require.NotZero(t, accountData.LastWAAUpdate)
	}

	assertExistsAndNonzeroWAAUpdate(false)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewSetValidation(targetAddress, targetPublic, []signature.PublicKey{newPublic}, []byte{}, 1, targetPrivate)
	resp := deliverTx(t, app, ca)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	assertExistsAndNonzeroWAAUpdate(true)
}
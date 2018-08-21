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

	// the address is invalid, but newClaimAccount doesn't validate this
	ca := NewClaimAccount(addr, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)

	// However, the resultant transaction must not be valid
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// what about an address which is valid but doesn't already exist?
	fakeTarget, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	ca = NewClaimAccount(fakeTarget, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)
	ctkBytes, err = tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimAccount(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestClaimAccountNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app := initAppClaimAccount(t)

	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{targetPublic}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimAccountUpdatesTransferKey(t *testing.T) {
	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
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
		require.Equal(t, newPublic.Bytes(), ad.TransferKeys[0].Bytes())
	})
}

func TestClaimAccountNoTransferKeys(t *testing.T) {
	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	app := initAppClaimAccount(t)
	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimAccountTooManyTransferKeys(t *testing.T) {
	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	ca := NewClaimAccount(targetAddress, targetPublic, newKeys, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
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
		ad.TransferKeys = []signature.PublicKey{existing}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
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
	modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
		ad.TransferKeys = []signature.PublicKey{existing1, existing2}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	ca := NewClaimAccount(targetAddress, targetPublic, []signature.PublicKey{newPublic}, targetPrivate)
	ctkBytes, err := tx.Marshal(&ca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

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
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestClaimChildAccountAddressFieldValidates(t *testing.T) {
	app, private := initAppTx(t)

	// Flip the bits of the last byte so the address is no longer correct.
	addrBytes := []byte(sourceAddress.String())
	addrBytes[len(addrBytes)-1] = ^addrBytes[len(addrBytes)-1]
	addrS := string(addrBytes)

	// Ensure that we didn't accidentally create a valid address.
	addr, err := address.Validate(addrS)
	require.Error(t, err)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// The address is invalid, but NewClaimChildAccount doesn't validate this.
	cca := NewClaimChildAccount(
		addr,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	// However, the resultant transaction must not be valid.
	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// What about an address which is valid but doesn't already exist?
	fakeSource, err := address.Generate(address.KindUser, addrBytes)
	require.NoError(t, err)
	cca = NewClaimChildAccount(
		fakeSource,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err = tx.Marshal(cca, TxIDs)
	require.NoError(t, err)
	resp = app.CheckTx(ctkBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimChildAccount(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// Apply the transaction as tendermint would.
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	dresp := app.DeliverTx(ctkBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// Ensure the child's settlement period matches the default from the system variable.
	t.Log(dresp.Log)
	child, _ := app.getAccount(childAddress)
	require.Equal(t, app.getDefaultSettlementDuration(), child.SettlementSettings.Period)
}

func TestClaimChildAccountSettlementPeriod(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	period := math.Duration(1234)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		period,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// Apply the transaction as tendermint would.
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	dresp := app.DeliverTx(ctkBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	// Ensure the child's settlement period matches what we set it to.
	t.Log(dresp.Log)
	child, _ := app.getAccount(childAddress)
	require.Equal(t, period, child.SettlementSettings.Period)
}

func TestClaimChildAccountNewTransferKeyNotEqualOwnershipKey(t *testing.T) {
	app, private := initAppTx(t)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{childPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidClaimChildAccountUpdatesTransferKey(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// Apply the transaction as tendermint would.
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	dresp := app.DeliverTx(ctkBytes)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	t.Log(dresp.Log)
	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))
	modify(t, childAddress.String(), app, func(ad *backing.AccountData) {
		require.Equal(t, newPublic.KeyBytes(), ad.ValidationKeys[0].KeyBytes())
	})
}

func TestClaimChildAccountNoValidationKeys(t *testing.T) {
	app, private := initAppTx(t)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimChildAccountTooManyValidationKeys(t *testing.T) {
	app, private := initAppTx(t)

	noKeys := backing.MaxKeysInAccount + 1
	newKeys := make([]signature.PublicKey, 0, noKeys)
	for i := 0; i < noKeys; i++ {
		key, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		newKeys = append(newKeys, key)
	}

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		newKeys,
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimChildAccountCannotHappenTwice(t *testing.T) {
	app, private := initAppTx(t)

	// Simulate the child account already having been claimed.
	existing, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	modify(t, childAddress.String(), app, func(ad *backing.AccountData) {
		ad.ValidationKeys = []signature.PublicKey{existing}
	})

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimChildAccountDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	modify(t, childAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		newPublic, _, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		cca := NewClaimChildAccount(
			sourceAddress,
			childAddress,
			childPublic,
			childSignature,
			childSettlementPeriod,
			[]signature.PublicKey{newPublic},
			[]byte{},
			1+uint64(i),
			private,
		)

		resp := deliverTxWithTxFee(t, app, cca)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestClaimChildAccountDoesntResetWAA(t *testing.T) {
	app, private := initAppTx(t)

	assertExistsAndNonzeroWAAUpdate := func(expectExists bool) {
		resp := app.Query(abci.RequestQuery{
			Path: query.AccountEndpoint,
			Data: []byte(childAddress.String()),
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

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		1,
		private,
	)

	resp := deliverTx(t, app, cca)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	assertExistsAndNonzeroWAAUpdate(true)
}

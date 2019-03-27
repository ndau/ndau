package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
)

func TestClaimChildAccountInvalidTargetAddress(t *testing.T) {
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
}

func TestClaimChildAccountInvalidChildAddress(t *testing.T) {
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
		sourceAddress,
		addr,
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
}

func TestClaimChildAccountNonExistentTargetAddress(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewClaimChildAccount(
		targetAddress,
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

	dresp := deliverTx(t, app, cca)
	t.Log(dresp.Log)

	// Ensure the child's settlement period matches the default from the system variable.
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

	dresp := deliverTx(t, app, cca)
	t.Log(dresp.Log)

	// Ensure the child's settlement period matches what we set it to.
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

	dresp := deliverTx(t, app, cca)
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

func TestClaimGrandchildAccount(t *testing.T) {
	app, sourceValidation := initAppTx(t)

	claimChild := func(
		parent address.Address,
		progenitor address.Address,
		parentPrivate signature.PrivateKey,
	) (address.Address, signature.PrivateKey) {
		childPublic, childPrivate, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		child, err := address.Generate(parent.Kind(), childPublic.KeyBytes())
		require.NoError(t, err)

		childSignature := childPrivate.Sign([]byte(child.String()))

		validationPublic, validationPrivate, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)

		parentAcct, _ := app.getAccount(parent)

		cca := NewClaimChildAccount(
			parent,
			child,
			childPublic,
			childSignature,
			childSettlementPeriod,
			[]signature.PublicKey{validationPublic},
			[]byte{},
			parentAcct.Sequence+1,
			parentPrivate,
		)

		// Make the progenitor an exchange account, to test more code paths.
		context := ddc(t).withExchangeAccount(progenitor)
		dresp, _ := deliverTxContext(t, app, cca, context)
		require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

		childAcct, exists := app.getAccount(child)
		require.True(t, exists)
		require.Equal(t, &parent, childAcct.Parent)
		require.Equal(t, &progenitor, childAcct.Progenitor)
		require.Equal(t, 1, len(cca.ChildValidationKeys))
		require.Equal(t, 1, len(childAcct.ValidationKeys))
		require.Equal(t,
			cca.ChildValidationKeys[0].KeyBytes(),
			childAcct.ValidationKeys[0].KeyBytes(),
		)

		// Since the progenitor was marked as an exchange account, so should
		// any descendant.
		// However, whether or not any account is an exchange account depends
		// on the state of the sysvars; these have been changed by our context.
		// We therefore need to get into that context.
		context.Within(app, func() {
			isExchangeAccount, err := app.GetState().(*backing.State).AccountHasAttribute(child, sv.AccountAttributeExchange)
			require.NoError(t, err)
			require.True(t, isExchangeAccount)
		})

		return child, validationPrivate
	}

	// Claim a child of the source account.
	child, childValidation := claimChild(sourceAddress, sourceAddress, sourceValidation)

	// Claim a child of the child (a grandchild of the source account).
	claimChild(child, sourceAddress, childValidation)
}

func TestClaimChildAccountInvalidValidationScript(t *testing.T) {
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
		[]byte{0x01},
		1,
		private,
	)

	ctkBytes, err := tx.Marshal(cca, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(ctkBytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimChildAccountInvalidChildSignature(t *testing.T) {
	app, private := initAppTx(t)

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	cca := NewClaimChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		private.Sign([]byte(childAddress.String())),
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

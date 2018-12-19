package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestValidDelegateTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	// d must be valid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestDelegateAccountValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	// make the account field invalid
	d.Target = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateDelegateValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	// make the account field invalid
	d.Node = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 0, private)

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	// flip a single bit in the signature
	sigBytes := d.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	d.Signatures[0] = *wrongSignature

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestDelegateChangesAppState(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's delegation node
	require.Equal(t, &nA, state.Accounts[source].DelegationNode)

	// we must have added the source to the node's delegation responsibilities
	require.Contains(t, state.Delegates, eaiNode)
	require.Contains(t, state.Delegates[eaiNode], source)
}

func TestDelegateRemovesPreviousDelegation(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// now create a new delegation transaction
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	d = NewDelegate(sA, dA, 2, private)
	resp = deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's delegation node
	require.Equal(t, &dA, state.Accounts[source].DelegationNode)

	// we must have added the source to dest's delegation responsibilities
	require.Contains(t, state.Delegates, dest)
	require.Contains(t, state.Delegates[dest], source)

	// we must have removed the source from eaiNode
	require.Contains(t, state.Delegates, eaiNode)
	require.NotContains(t, state.Delegates[eaiNode], source)
}

func TestDelegateDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewDelegate(sA, nA, 1+uint64(i), private)

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

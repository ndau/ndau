package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

// Private key:    e8f080d6f39b0942217a55a4e239cc59b6dfbc48bc3d5e0abebc7da0bf055f57d17516973974aced03ca0ebef33b3798719c596b01a065a0de74e999670e1be5
// Public key:     d17516973974aced03ca0ebef33b3798719c596b01a065a0de74e999670e1be5
const eaiNode = "ndamb84tesvp54vhc63257wifr34zfvyffvi9utqrkruneai"

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
	d.Account = address.Address{}
	d.Signature = private.Sign(d.signableBytes())

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
	d.Delegate = address.Address{}
	d.Signature = private.Sign(d.signableBytes())

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
	sigBytes := d.Signature.Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	d.Signature = *wrongSignature

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

	resp := deliverTr(t, app, d)
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

	resp := deliverTr(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// now create a new delegation transaction
	dA, err := address.Validate(dest)
	require.NoError(t, err)
	d = NewDelegate(sA, dA, 2, private)
	resp = deliverTr(t, app, d)
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

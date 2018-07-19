package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

func initAppComputeEAI(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)

	// delegate source to eaiNode
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, private)
	resp := deliverTr(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// create a keypair for the node
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	// assign this keypair
	modify(t, eaiNode, app, func(data *backing.AccountData) {
		data.TransferKey = &public
	})
	return app, private
}

func TestValidComputeEAITxIsValid(t *testing.T) {
	app, private := initAppComputeEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewComputeEAI(nA, 1, private)
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestComputeEAINodeValidates(t *testing.T) {
	app, private := initAppComputeEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewComputeEAI(nA, 2, private)

	// make the node field invalid
	compute.Node = address.Address{}
	compute.Signature = private.Sign(compute.signableBytes())

	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestComputeEAISequenceValidates(t *testing.T) {
	app, private := initAppComputeEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewComputeEAI(nA, 0, private)
	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestComputeEAISignatureValidates(t *testing.T) {
	app, private := initAppComputeEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewComputeEAI(nA, 0, private)

	// flip a single bit in the signature
	sigBytes := compute.Signature.Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	compute.Signature = *wrongSignature

	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestComputeEAIChangesAppState(t *testing.T) {
	app, private := initAppComputeEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewComputeEAI(nA, 1, private)

	state := app.GetState().(*backing.State)
	sourceInitial := state.Accounts[source].Balance

	blockTime := math.Timestamp(45 * math.Day)
	bt := constants.Epoch.Add(math.Duration(blockTime).TimeDuration())
	resp := deliverTrAt(t, app, compute, bt.Unix())
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	state = app.GetState().(*backing.State)
	acct := state.Accounts[source]
	t.Log(acct.Balance)
	// here, we don't bother testing _how much_ eai is applied: we have to
	// trust that the ndaumath library is well tested. Instead, we just test
	// that _more than 0_ eai is applied.
	require.Equal(t, -1, sourceInitial.Compare(acct.Balance))
	// n.b. These two times are equal in this case, but they are sometimes
	// distinct. A transfer needs to update WAA but not EAI, so they can
	// be different.
	require.Equal(t, blockTime, acct.LastEAIUpdate)
	require.Equal(t, blockTime, acct.LastWAAUpdate)
}

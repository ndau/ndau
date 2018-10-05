package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

func initAppCreditEAI(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)

	// delegate source to eaiNode
	sA, err := address.Validate(source)
	require.NoError(t, err)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	d := NewDelegate(sA, nA, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, d)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.LastEAIUpdate = 0
		ad.LastWAAUpdate = 0
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// create a keypair for the node
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	// assign this keypair
	modify(t, eaiNode, app, func(data *backing.AccountData) {
		data.ValidationKeys = []signature.PublicKey{public}
	})
	return app, private
}

func TestValidCreditEAITxIsValid(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 1, []signature.PrivateKey{private})
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestCreditEAINodeValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 2, []signature.PrivateKey{private})

	// make the node field invalid
	compute.Node = address.Address{}
	compute.Signatures = []signature.Signature{private.Sign(compute.SignableBytes())}

	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAISequenceValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 0, []signature.PrivateKey{private})
	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAISignatureValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 0, []signature.PrivateKey{private})

	// flip a single bit in the signature
	sigBytes := compute.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	compute.Signatures[0] = *wrongSignature

	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAIChangesAppState(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 1, []signature.PrivateKey{private})

	state := app.GetState().(*backing.State)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	acct, _ := state.GetAccount(sA, app.blockTime)
	sourceInitial := acct.Balance

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTrAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	state = app.GetState().(*backing.State)
	acct, _ = state.GetAccount(sA, app.blockTime)
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

func TestCreditEAIWithRewardsTargetChangesAppState(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 1, []signature.PrivateKey{private})

	sA, err := address.Validate(source)
	require.NoError(t, err)
	state := app.GetState().(*backing.State)
	sAcct, _ := state.GetAccount(sA, app.blockTime)
	sourceInitial := sAcct.Balance

	dA, err := address.Validate(dest)
	require.NoError(t, err)
	// verify that the dest account has nothing currently in it
	dAcct, _ := state.GetAccount(dA, app.blockTime)
	require.Equal(t, math.Ndau(0), dAcct.Balance)
	// have the source acct send rewards to the dest acct
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &dA
	})

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTrAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	state = app.GetState().(*backing.State)
	sAcct, _ = state.GetAccount(sA, app.blockTime)
	dAcct, dExists := state.GetAccount(dA, app.blockTime)
	t.Log("src:  ", sAcct.Balance)
	t.Log("dest: ", dAcct.Balance)
	require.True(t, dExists)
	// the source account must not be changed
	require.Equal(t, sourceInitial, sAcct.Balance)
	// the dest acct must now have a non-0 balance
	require.NotEqual(t, math.Ndau(0), dAcct.Balance)
}

func TestCreditEAIWithNotifiedRewardsTargetIsAllowed(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	compute := NewCreditEAI(nA, 1, []signature.PrivateKey{private})

	sA, err := address.Validate(source)
	require.NoError(t, err)
	state := app.GetState().(*backing.State)
	sAcct, _ := state.GetAccount(sA, app.blockTime)
	sourceInitial := sAcct.Balance

	dA, err := address.Validate(dest)
	require.NoError(t, err)
	// verify that the dest account has nothing currently in it
	dAcct, _ := state.GetAccount(dA, app.blockTime)
	require.Equal(t, math.Ndau(0), dAcct.Balance)
	// have the source acct send rewards to the dest acct
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &dA
	})
	modify(t, dest, app, func(ad *backing.AccountData) {
		uo := math.Timestamp(1 * math.Year)
		ad.Lock = &backing.Lock{
			NoticePeriod: math.Duration(1 * math.Year),
			UnlocksOn:    &uo,
		}
	})

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTrAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that eai was deposited despite the dest acct being notified
	state = app.GetState().(*backing.State)
	sAcct, _ = state.GetAccount(sA, app.blockTime)
	dAcct, _ = state.GetAccount(dA, app.blockTime)
	t.Log("src:  ", sAcct.Balance)
	t.Log("dest: ", dAcct.Balance)
	// the source account must not be changed
	require.Equal(t, sourceInitial, sAcct.Balance)
	// the dest acct must have had some EAI credited
	require.NotEqual(t, math.Ndau(0), dAcct.Balance)
}

func TestCreditEAIDeductsTxFee(t *testing.T) {
	app, private := initAppCreditEAI(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewCreditEAI(nA, 1+uint64(i), []signature.PrivateKey{private})

		resp := deliverTrWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

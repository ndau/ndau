package ndau

import (
	"sort"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func initAppCreditEAI(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)

	// delegate source to eaiNode
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)
	resp := deliverTx(t, app, d)
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
	compute := NewCreditEAI(nodeAddress, 1, private)
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestCreditEAINodeValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 2, private)

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
	compute := NewCreditEAI(nodeAddress, 0, private)
	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAISignatureValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 0, private)

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
	compute := NewCreditEAI(nodeAddress, 1, private)

	state := app.GetState().(*backing.State)
	acct, _ := state.GetAccount(sourceAddress, app.blockTime)
	sourceInitial := acct.Balance

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTxAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	state = app.GetState().(*backing.State)
	acct, _ = state.GetAccount(sourceAddress, app.blockTime)
	t.Log(acct.Balance)
	// here, we don't bother testing _how much_ eai is applied: we have to
	// trust that the ndaumath library is well tested. Instead, we just test
	// that _more than 0_ eai is applied.
	require.Equal(t, -1, sourceInitial.Compare(acct.Balance))
	// n.b. These two times are equal in this case, but they are sometimes
	// distinct. A transfer needs to update WAA but not EAI, so they can
	// be different.
	require.Equal(t, blockTime, acct.LastEAIUpdate)
	// EAI does not update WAA when it's delivered to the same account
	require.NotEqual(t, blockTime, acct.LastWAAUpdate)
}

func TestCreditEAIWithRewardsTargetChangesAppState(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	state := app.GetState().(*backing.State)
	sAcct, _ := state.GetAccount(sourceAddress, app.blockTime)
	sourceInitial := sAcct.Balance

	// verify that the dest account has nothing currently in it
	dAcct, _ := state.GetAccount(destAddress, app.blockTime)
	require.Equal(t, math.Ndau(0), dAcct.Balance)
	// have the source acct send rewards to the dest acct
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &destAddress
	})

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTxAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	state = app.GetState().(*backing.State)
	sAcct, _ = state.GetAccount(sourceAddress, app.blockTime)
	dAcct, dExists := state.GetAccount(destAddress, app.blockTime)
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
	compute := NewCreditEAI(nodeAddress, 1, private)

	state := app.GetState().(*backing.State)
	sAcct, _ := state.GetAccount(sourceAddress, app.blockTime)
	sourceInitial := sAcct.Balance

	// verify that the dest account has nothing currently in it
	dAcct, _ := state.GetAccount(destAddress, app.blockTime)
	require.Equal(t, math.Ndau(0), dAcct.Balance)
	// have the source acct send rewards to the dest acct
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.RewardsTarget = &destAddress
	})
	modify(t, dest, app, func(ad *backing.AccountData) {
		uo := math.Timestamp(1 * math.Year)
		ad.Lock = backing.NewLock(1*math.Year, eai.DefaultLockBonusEAI)
		ad.Lock.UnlocksOn = &uo
	})

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTxAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that eai was deposited despite the dest acct being notified
	state = app.GetState().(*backing.State)
	sAcct, _ = state.GetAccount(sourceAddress, app.blockTime)
	dAcct, _ = state.GetAccount(destAddress, app.blockTime)
	t.Log("src:  ", sAcct.Balance)
	t.Log("dest: ", dAcct.Balance)
	// the source account must not be changed
	require.Equal(t, sourceInitial, sAcct.Balance)
	// the dest acct must have had some EAI credited
	require.NotEqual(t, math.Ndau(0), dAcct.Balance)
}

func TestCreditEAIDeductsTxFee(t *testing.T) {
	app, private := initAppCreditEAI(t)
	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewCreditEAI(nodeAddress, 1+uint64(i), private)

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

// the problem we're seeing is that the iteration of accounts in the CreditEAI
// transaction is happening in random order. In the situation where an account
// appears in the iteration more than once (as the target of one or more EAIs
// AND as an account earning EAI itself), the result is order-dependent.
// Consequently, two nodes can process the same tx and get different results.
//
// We could, fairly easily, sort the list and make sure that all nodes get the
// same result. But in a case like this:
//
// A 1000 -> B
// B 100
// C 100
// D 1000 -> C
//
// B and C will end up with different results. That's not acceptable.
func TestCreditEAIIsDeterministic(t *testing.T) {
	// set up accounts
	type account struct {
		address address.Address
		private signature.PrivateKey
		public  signature.PublicKey
	}

	makeAccount := func() account {
		public, private, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		addr, err := address.Generate(address.KindUser, public.KeyBytes())
		require.NoError(t, err)
		return account{
			address: addr,
			private: private,
			public:  public,
		}
	}

	accts := make([]account, 5)
	for i := range accts {
		accts[i] = makeAccount()
	}

	// part of the point of the test is to exclude "solutions" in which
	// two otherwise-equivalent account pairs experience different behavior
	// depending on the relative ordering of their addresses. That sort of
	// thing would be deterministic, but it wouldn't be fair. Therefore,
	// we need to sort the accounts by address to ensure that we have one
	// account whose rewards are redirected to another one after it,
	// and one whose rewards are redirected to another one before it.
	// Therefore, we need to sort these accounts.
	sort.Slice(accts, func(i, j int) bool {
		return accts[i].address.String() < accts[j].address.String()
	})

	a := accts[0]
	b := accts[1]
	c := accts[2]
	d := accts[3]
	node := accts[4]

	// set up app preconditions
	app, _ := initApp(t)

	var txs []metatx.Transactable

	delegate := func(from account) {
		txs = append(txs,
			NewDelegate(from.address, node.address, 1, from.private),
		)
	}

	redirect := func(from, to account) {
		txs = append(txs,
			NewSetRewardsDestination(
				from.address,
				to.address,
				2,
				from.private,
			),
		)
	}

	setup := func(acct account, balance math.Ndau, redirectTo *account) {
		modify(t, acct.address.String(), app, func(ad *backing.AccountData) {
			ad.LastEAIUpdate = 0
			ad.LastWAAUpdate = 0
			ad.WeightedAverageAge = 0
			ad.Balance = balance * constants.NapuPerNdau
			// normally it's enforced that a validation key can't match the
			// ownership key for the account, but it's not important for the
			// behavior under test, and this is simpler instead of having
			// to generate even more keys
			ad.ValidationKeys = []signature.PublicKey{acct.public}
		})
		if acct.address.String() != node.address.String() {
			delegate(acct)
		}
		if redirectTo != nil {
			redirect(acct, *redirectTo)
		}
	}

	setup(a, 1000, &b)
	setup(b, 100, nil)
	setup(c, 100, nil)
	setup(d, 1000, &c)
	setup(node, 10, nil)

	resps, _ := deliverTxsContext(t, app, txs, ddc(t))
	for _, resp := range resps {
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	}

	t.Logf("a:    %s", a.address)
	t.Logf("b:    %s", b.address)
	t.Logf("c:    %s", c.address)
	t.Logf("d:    %s", d.address)
	t.Logf("node: %s", node.address)

	equiv := func(a, b backing.AccountData) {
		require.Equal(t, a.Balance, b.Balance)
		require.Equal(t, a.LastEAIUpdate, b.LastEAIUpdate)
		require.Equal(t, a.LastWAAUpdate, b.LastWAAUpdate)
		require.Equal(t, a.WeightedAverageAge, b.WeightedAverageAge)
	}

	checkState := func() {
		// given the exact same circumstances, each account data pair must
		// be identical
		state := app.GetState().(*backing.State)

		A, _ := state.GetAccount(a.address, app.blockTime)
		B, _ := state.GetAccount(b.address, app.blockTime)
		C, _ := state.GetAccount(c.address, app.blockTime)
		D, _ := state.GetAccount(d.address, app.blockTime)

		equiv(A, D)
		equiv(B, C)
	}

	// we must have set up the initial state of the accounts identically
	checkState()

	// perform tests
	// note: we do _not_ wish to run each iteration here as an independent
	// subtest, which is a little unusual. However, the intent here is that
	// we don't actually vary the conditions within each test instance; we just
	// want some measure of reassurance that we're not just getting lucky
	// with random map iteration order.
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)
	for i := uint64(0); i < 128; i++ {
		// perform a CreditEAI tx
		tx := NewCreditEAI(node.address, i+5, node.private)

		resp := deliverTxAt(
			t,
			app,
			tx,
			now.Add(math.Duration(i*math.Month)),
		)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))

		// state must still be identical here
		checkState()
	}
}

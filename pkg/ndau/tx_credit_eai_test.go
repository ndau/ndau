package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"sort"
	"testing"
	"time"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	srch "github.com/ndau/ndau/pkg/ndau/search"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/eai"
	"github.com/ndau/ndaumath/pkg/signature"
	"github.com/ndau/ndaumath/pkg/signed"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppCreditEAI(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppDelegate(t)

	// delegate source to eaiNode
	d := NewDelegate(sourceAddress, nodeAddress, 1, private)
	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	ensureRecent(t, app, sourceAddress.String())
	ensureRecent(t, app, nodeAddress.String())
	ensureRecent(t, app, targetAddress.String())

	// create a keypair for the node
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	// assign this keypair
	modify(t, eaiNode, app, func(data *backing.AccountData) {
		data.ValidationKeys = []signature.PublicKey{public}
	})
	return app, private
}

func initAppCreditEAIWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	*App, signature.PrivateKey,
) {
	app, private := initAppDelegate(t)
	return app, private
}

func TestValidCreditEAITxIsValid(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
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
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAISequenceValidates(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 0, private)
	// compute must be invalid
	bytes, err := tx.Marshal(compute, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
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
	resp := app.CheckTx(abci.RequestCheckTx{Tx: bytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAIChangesAppState(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	acct, _ := app.getAccount(sourceAddress)
	sourceInitial := acct.Balance

	blockTime := math.Timestamp(45 * math.Day)
	resp := deliverTxAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	acct, _ = app.getAccount(sourceAddress)
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

func TestCreditEAIHandlesExchangeAccounts(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	acct, _ := app.getAccount(sourceAddress)
	sourceInitial := acct.Balance

	rate, err := app.calculateExchangeEAIRate(acct)
	require.NoError(t, err)
	t.Log("rate:", rate.String())

	// expected EAI = BALANCE * (e^(RATE*TIME) - 1)
	// however, as TIME == 1, we can exclude it from our calculations
	// e ^ RATE
	expectedEAI, err := signed.ExpFrac(int64(rate), constants.RateDenominator)
	require.NoError(t, err)
	t.Log("e^RATE =", expectedEAI)
	// x-1
	expectedEAI -= constants.RateDenominator
	t.Log("(e^RATE)-1 =", expectedEAI)
	// BALANCE * x
	t.Log("sourceInitial =", sourceInitial.String())
	expectedEAI, err = signed.MulDiv(int64(sourceInitial), expectedEAI, constants.RateDenominator)
	require.NoError(t, err)
	t.Log("expectedEAI =", math.Ndau(expectedEAI).String())
	// Subtract off the 15% EAI fees.
	expectedEAI, err = signed.MulDiv(
		int64(expectedEAI),
		int64(eai.RateFromPercent(85)),
		constants.RateDenominator)
	require.NoError(t, err)
	t.Log("expectedEAI less fees =", math.Ndau(expectedEAI).String())

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.LastEAIUpdate = 0
		ad.LastWAAUpdate = 0
		ad.UncreditedEAI = 0
	})

	context := ddc(t).
		at(math.Timestamp(1 * math.Year)).
		withExchangeAccount(sourceAddress).
		with(func(svs map[string][]byte) {
			delete(svs, sv.EAIOvertime)
		})
	resp, _ := deliverTxContext(t, app, compute, context)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	acct, _ = app.getAccount(sourceAddress)
	t.Log(acct.Balance)
	require.Equal(t, sourceInitial+math.Ndau(expectedEAI), acct.Balance)
}

func TestCreditEAIUpdatesCurrencySeat(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	modify(t, sourceAddress.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 999 * constants.QuantaPerUnit
		ad.CurrencySeatDate = nil
	})

	acct, _ := app.getAccount(sourceAddress)

	// we want enough time to earn some ndau
	blockTime := math.Timestamp(90 * math.Day)
	resp := deliverTxAt(t, app, compute, blockTime)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	acct, _ = app.getAccount(sourceAddress)
	t.Log("BALANCE: ", acct.Balance)
	require.True(t, acct.Balance > 1000*constants.QuantaPerUnit)
	require.NotNil(t, acct.CurrencySeatDate)
}

func TestCreditEAIWithRewardsTargetChangesAppState(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	sAcct, _ := app.getAccount(sourceAddress)
	sourceInitial := sAcct.Balance

	// verify that the dest account has nothing currently in it
	dAcct, _ := app.getAccount(destAddress)
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
	sAcct, _ = app.getAccount(sourceAddress)
	dAcct, dExists := app.getAccount(destAddress)
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

	sAcct, _ := app.getAccount(sourceAddress)
	sourceInitial := sAcct.Balance

	// verify that the dest account has nothing currently in it
	dAcct, _ := app.getAccount(destAddress)
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
	sAcct, _ = app.getAccount(sourceAddress)
	dAcct, _ = app.getAccount(destAddress)
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

	var txs []tx.Transactable

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
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node.address.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})

	resps, _ := deliverTxsContext(t, app, txs, ddc(t).at(math.Timestamp(math.Year)))
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
		A, _ := app.getAccount(a.address)
		B, _ := app.getAccount(b.address)
		C, _ := app.getAccount(c.address)
		D, _ := app.getAccount(d.address)

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
	ts := math.Timestamp(math.Year + math.Month)
	for i := uint64(0); i < 128; i++ {
		// perform a CreditEAI tx
		tx := NewCreditEAI(node.address, i+5, node.private)

		resp := deliverTxAt(
			t,
			app,
			tx,
			ts,
		)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))

		// state must still be identical here
		checkState()
	}
}

// like the previous, but don't pre-create B and C
func TestCreditEAIIsDeterministic2(t *testing.T) {
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

	var txs []tx.Transactable

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
	setup(d, 1000, &c)
	setup(node, 10, nil)
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		st.Nodes[node.address.String()] = backing.Node{
			Active: true,
		}
		return st, nil
	})

	resps, _ := deliverTxsContext(t, app, txs, ddc(t).at(math.Timestamp(math.Year)))
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
		A, _ := app.getAccount(a.address)
		B, _ := app.getAccount(b.address)
		C, _ := app.getAccount(c.address)
		D, _ := app.getAccount(d.address)

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
	ts := math.Timestamp(math.Year + math.Month)
	for i := uint64(0); i < 128; i++ {
		// perform a CreditEAI tx
		tx := NewCreditEAI(node.address, i+5, node.private)

		resp := deliverTxAt(
			t,
			app,
			tx,
			ts,
		)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))

		// state must still be identical here
		checkState()
	}
}

func TestCreditEAIClearsUncreditedEAI(t *testing.T) {
	app, private := initAppCreditEAI(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.UncreditedEAI = 12345
	})
	tx := NewCreditEAI(nodeAddress, 1, private)
	resp := deliverTxAt(t, app, tx, 45*math.Day)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	sourceData := app.GetState().(*backing.State).Accounts[source]
	require.Zero(t, sourceData.UncreditedEAI)
}

func TestCreditEAICanOnlyBeSubmittedByActiveNode(t *testing.T) {
	app, private := initAppCreditEAI(t)
	// ensure node is not active, for testing purposes
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)
		node := st.Nodes[nodeAddress.String()]
		node.Active = false
		st.Nodes[eaiNode] = node
		return st, nil
	})
	tx := NewCreditEAI(nodeAddress, 1, private)
	resp := deliverTxAt(t, app, tx, 45*math.Day)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestCreditEAIUsesOvertimeAppropriately(t *testing.T) {
	app, private := initAppCreditEAI(t)
	compute := NewCreditEAI(nodeAddress, 1, private)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// source has 10000 ndau exactly
	// EAI overtime limit is 30 days
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.UncreditedEAI = 0
		ad.LastEAIUpdate = ts - (45 * math.Day)
		ad.LastWAAUpdate = ts - (45 * math.Day)
		ad.WeightedAverageAge = 0
	})

	resp := deliverTxAt(t, app, compute, ts)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// require that a positive EAI was applied
	acct, _ := app.getAccount(sourceAddress)
	require.Nil(t, acct.Lock)
	t.Log(acct.Balance)

	// we must have applied exactly 30 days' worth of EAI using the standard table,
	// despite the 45 days which have accrued since the last update.
	// How much is that, anyway?
	eaiAward, err := eai.Calculate(
		10000*constants.NapuPerNdau, 45*math.Day, 15*math.Day,
		45*math.Day, nil,
		eai.DefaultUnlockedEAI, true,
	)
	require.NoError(t, err)
	feeTable := new(sv.EAIFeeTable)
	err = app.System(sv.EAIFeeTableName, feeTable)
	require.NoError(t, err)
	awardPerNdau := math.Ndau(constants.QuantaPerUnit)
	for _, fee := range *feeTable {
		awardPerNdau -= fee.Fee
	}
	reducedAward, err := signed.MulDiv(
		int64(eaiAward),
		int64(awardPerNdau),
		constants.QuantaPerUnit,
	)
	require.NoError(t, err)
	expect := math.Ndau((10000 * constants.NapuPerNdau) + reducedAward)

	require.Equal(t, expect, acct.Balance)
}

func TestCreditEAIRetainsPendingLock(t *testing.T) {
	app, private := initAppCreditEAI(t)

	// set up the source account such that it is delegated to the node
	err := app.UpdateStateImmediately(app.Delegate(sourceAddress, nodeAddress))
	require.NoError(t, err)
	// ensure source is locked and has not yet unlocked
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = backing.NewLock(math.Year, eai.DefaultLockBonusEAI)
		ad.Lock.Notify(0, 0)
	})

	// source is locked for a year at time 0
	// to test whether the lock is retained, we deliver the credit after less than a year

	tx := NewCreditEAI(nodeAddress, 1, private)
	resp := deliverTxAt(t, app, tx, 6*math.Month)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// ensure the account is still locked
	ad, _ := app.getAccount(sourceAddress)
	require.NotNil(t, ad.Lock)
}

func TestCreditEAIClearsCompletedLock(t *testing.T) {
	app, private := initAppCreditEAI(t)

	// set up the source account such that it is delegated to the node
	err := app.UpdateStateImmediately(app.Delegate(sourceAddress, nodeAddress))
	require.NoError(t, err)
	// ensure source is locked and has not yet unlocked
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = backing.NewLock(math.Year, eai.DefaultLockBonusEAI)
		ad.Lock.Notify(0, 0)
	})

	// source is locked for a year at time 0
	// to test whether the lock is retained, we deliver the credit after more than a year

	tx := NewCreditEAI(nodeAddress, 1, private)
	resp := deliverTxAt(t, app, tx, 18*math.Month)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// ensure the account is still locked
	ad, _ := app.getAccount(sourceAddress)
	require.Nil(t, ad.Lock)
}

func TestRecalculateWAA(t *testing.T) {
	withRedis(t, func(port string) {
		// Create the app and tx factory.
		app, private := initAppCreditEAIWithIndex(t, "localhost:"+port, 0)

		if app.config.Features == nil {
			app.config.Features = make(map[string]uint64)
		}
		// WAA recalculation is only necessary because of a feature-gated
		// fix we'd previously put in, so let's enable that feature
		app.config.Features["ApplyUncreditedEAI"] = 0

		search := app.GetSearch().(*srch.Client)

		// Ensure Redis is empty.
		err := search.FlushDB()
		require.NoError(t, err)

		// Test initial indexing.
		var insertCount int
		_, insertCount, err = search.IndexBlockchain(app.GetDB(), app.GetDS())
		require.NoError(t, err)

		state := app.GetState().(*backing.State)
		numSysvars := len(state.Sysvars)

		// The sysvars should all be inserted
		require.GreaterOrEqual(t, insertCount, numSysvars)

		// we need the auto-incrementing height from a persistent delivery
		// context
		dc := ddc(t).atHeight(1000)

		// set up an account with a bunch of outbound transfers, triggering
		// the WAA bug
		makeacct := func(initialize bool) (address.Address, signature.PrivateKey) {
			public, private, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)
			addr, err := address.Generate(address.KindUser, public.KeyBytes())
			require.NoError(t, err)

			if initialize {
				// just modifying the account with both balance and validation
				// keys is insufficient, because the account's creation has
				// to appear within the postgres instance. It therefore needs
				// a real tx.

				// first, set up the source account
				modify(t, source, app, func(ad *backing.AccountData) {
					ad.ValidationKeys = []signature.PublicKey{public}
					ad.Balance = 1500 * constants.NapuPerNdau
					ad.Sequence = 1
				})
				ensureRecent(t, app, source)

				// now create 'our' account by transfering into it from the source
				// note that we set the age to a year ago and decr the height
				resp, _ := deliverTxContext(
					t, app,
					NewTransfer(
						sourceAddress, addr,
						1000*constants.NapuPerNdau,
						2,
						private,
					),
					dc.at(dc.ts-math.Year).atHeight(dc.blockHeight-800),
				)
				require.Equal(t, code.OK, code.ReturnCode(resp.Code))
				// incr dc, because it (intentionally) doesn't auto-propagate in this case
				dc.incr()

				// the account exists, so let's shove in the public key
				modify(t, addr.String(), app, func(ad *backing.AccountData) {
					ad.ValidationKeys = []signature.PublicKey{public}
				})
			}

			return addr, private
		}

		// setup
		performSetup := func() address.Address {
			// part of setup is ensuring that the WAA update feature is disabled
			// by default. Specific tests can opt into a lower height gate
			// for this feature.
			app.config.Features["UpdateWAAUpdateDateInDetails"] = 1000000

			addr, private := makeacct(true)

			// the bug happens when a bunch of small transfers happen in quick
			// succession
			for i := uint64(0); i < 100; i++ {
				dest, _ := makeacct(false)
				resp, _ := deliverTxContext(t, app,
					NewTransfer(addr, dest, 1*constants.NapuPerNdau, i+10, private),
					dc,
				)
				require.Equal(t, code.OK, code.ReturnCode(resp.Code))
			}

			return addr
		}

		// test that setup triggered the bug
		t.Run("setup", func(t *testing.T) {
			addr := performSetup()
			acctData, _ := app.getAccount(addr)
			require.Greater(t, int64(acctData.WeightedAverageAge), int64(math.Year))
		})

		// tests
		t.Run("recalc without feature gate is noop", func(t *testing.T) {
			addr := performSetup()
			acctData, _ := app.getAccount(addr)
			oldWAA := acctData.WeightedAverageAge

			tx := NewCreditEAI(nodeAddress, 1, private)
			err = app.UpdateState(app.recalculateWAAs(tx))
			require.NoError(t, err)
			acctData, _ = app.getAccount(addr)
			require.Equal(t, oldWAA, acctData.WeightedAverageAge)
		})

		t.Run("recalc with feature gate", func(t *testing.T) {
			addr := performSetup()
			acctData, _ := app.getAccount(addr)
			oldWAA := acctData.WeightedAverageAge

			// Let's assume that the CreditEAI which calls recalculateWAAs
			// happens on the block subsequent to the final setup.
			// Let's further assume that that block's time is at least 1
			// microsecond after the previous block's. Both of these feel like
			// fairly safe assumptions.
			//
			// In that case, the index should contain the block height and time
			// for the block after the previous transaction, and the feature
			// gate needs to be configured appropriately. We can make that
			// happen.
			app.config.Features["UpdateWAAUpdateDateInDetails"] = dc.blockHeight - 1
			dc.beginBlock(app)

			tx := NewCreditEAI(nodeAddress, 1, private)
			err = app.UpdateState(app.recalculateWAAs(tx))
			require.NoError(t, err)
			acctData, _ = app.getAccount(addr)
			t.Log("old WAA:", oldWAA.String())
			t.Log("current WAA:", acctData.WeightedAverageAge.String())
			require.Greater(t, oldWAA, acctData.WeightedAverageAge)
			// we know what the WAA should be, in this case: setup ensures
			// the account is one year old and has only outbound transfers.
			// Given that setup, the WAA must be exactly one year.
			require.Equal(t, math.Duration(math.Year+(100*math.Microsecond)), acctData.WeightedAverageAge)
		})
	})
}

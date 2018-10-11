package ndau

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func initAppTx(t *testing.T) (*App, signature.PrivateKey) {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// generate the transfer key so we can transfer from it
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	modifySource(t, app, func(acct *backing.AccountData) {
		// initialize the source address with a bunch of ndau
		acct.Balance = math.Ndau(10000 * constants.QuantaPerUnit)
		acct.ValidationKeys = []signature.PublicKey{public}
	})

	return app, private
}

// generate an app with an account with a bunch of escrowed transactions
//
// returns that account's private key, and a timestamp after which all escrows
// should be valid
//
// It is guaranteed that all escrows expire in the interval (timestamp - 1 day : timestamp)
func initAppSettlement(t *testing.T) (*App, signature.PrivateKey, math.Timestamp) {
	app, _ := initAppTx(t)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// generate the transfer key so we can transfer from the escrowed acct
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	const qtyEscrows = 10

	modify(t, settled, app, func(acct *backing.AccountData) {
		// initialize the address with a bunch of ndau
		// incoming funds are added to the balance and the settlements;
		// it's just that the available balance is reduced by the sum
		// of the uncleared settlements
		for i := 1; i < qtyEscrows; i++ {
			acct.Balance += math.Ndau(i * constants.QuantaPerUnit)
			acct.Settlements = append(acct.Settlements, backing.Settlement{
				Qty:    math.Ndau(i * constants.QuantaPerUnit),
				Expiry: ts.Sub(math.Duration(i)),
			})
		}
		acct.ValidationKeys = []signature.PublicKey{public}
	})

	// add 1 second to the timestamp to get past unix time rounding errors
	tn := constants.Epoch.Add(time.Duration(int64(ts)) * time.Microsecond)
	tn = tn.Add(time.Duration(1 * time.Second))

	// update the app's cached timestamp
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: tn.Unix(),
	}})
	app.EndBlock(abci.RequestEndBlock{})

	return app, private, ts
}

// update the source account
func modifySource(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, source, app, f)
}

// update the dest account
func modifyDest(t *testing.T, app *App, f func(*backing.AccountData)) {
	modify(t, dest, app, f)
}

func generateTransfer(t *testing.T, qty int64, seq uint64, keys []signature.PrivateKey) *Transfer {
	s, err := address.Validate(source)
	require.NoError(t, err)
	d, err := address.Validate(dest)
	require.NoError(t, err)
	tr, err := NewTransfer(
		s, d,
		math.Ndau(qty*constants.QuantaPerUnit),
		seq, keys,
	)
	require.NoError(t, err)
	return tr
}

func TestTransfersWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []int64{0, -1, -2} {
		tr := generateTransfer(t, negQty, uint64(idx+1), []signature.PrivateKey{private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestTransfersFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = &backing.Lock{
			NoticePeriod: 90 * math.Day,
		}
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersFromLockedButExpiredAddressesAreValid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		twoDaysAgo := now.Sub(math.Duration(2 * math.Day))
		acct.Lock = &backing.Lock{
			NoticePeriod: math.Duration(1 * math.Day),
			UnlocksOn:    &twoDaysAgo,
		}
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTransfersFromNotifiedAddressesAreInvalid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		tomorrow := now.Add(math.Duration(1 * math.Day))
		acct.Lock = &backing.Lock{
			NoticePeriod: math.Duration(1 * math.Day),
			UnlocksOn:    &tomorrow,
		}
	})

	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransfersUpdateDestWAA(t *testing.T) {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)

	modifyDest(t, app, func(acct *backing.AccountData) {
		acct.Balance = 100 * constants.QuantaPerUnit
		acct.LastWAAUpdate = timestamp.Sub(math.Duration(30 * math.Day))
	})

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTrAt(t, app, tr, timestamp)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// because we're doing integer math on times here, we won't get
	// exact results. The results are _deterministic_, just not
	// _exact_. As such, we need to define success in terms of
	// error margins.
	//
	// Given that we're constrained by tendermint limitations to
	// block times at a resolution of 1 second anyway, it makes sense
	// to require that we calculate the correct second.
	const maxEpsilon = int64(1000) * math.Millisecond
	var epsilon int64
	expect := int64(20 * math.Day)
	// not actually modifying the dest here; this is just the
	// fastest way to get access to the account data
	modifyDest(t, app, func(acct *backing.AccountData) {
		epsilon = expect - int64(acct.WeightedAverageAge)
	})
	if epsilon < 0 {
		epsilon = -epsilon
	}
	require.True(
		t, epsilon < maxEpsilon,
		"must be true: epsilon < maxEpsilon",
		"epsilon", epsilon,
		"max epsilon", maxEpsilon,
	)
}

func TestTransfersUpdateDestLastWAAUpdate(t *testing.T) {
	timestamp, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// truncate timestamp to the nearest second:
	// we assume that unix time is what tendermint uses,
	// and unix time is truncated to seconds
	timestamp = math.Timestamp(int64(timestamp) - (int64(timestamp) % math.Second))

	app, private := initAppTx(t)

	modifyDest(t, app, func(acct *backing.AccountData) {
		acct.Balance = 100 * constants.QuantaPerUnit
		acct.LastWAAUpdate = timestamp.Sub(math.Duration(30 * math.Day))
	})

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTrAt(t, app, tr, timestamp)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// not actually modifying the dest here; this is just the
	// fastest way to get access to the account data
	modifyDest(t, app, func(acct *backing.AccountData) {
		require.Equal(t, timestamp, acct.LastWAAUpdate)
	})
}

func TestTransfersDeductBalanceFromSource(t *testing.T) {
	app, private := initAppTx(t)

	var initialSourceNdau int64
	modifySource(t, app, func(src *backing.AccountData) {
		initialSourceNdau = int64(src.Balance)
	})

	const deltaNapu = 50 * constants.QuantaPerUnit

	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		require.Equal(t, initialSourceNdau-deltaNapu, int64(src.Balance))
	})
}

func TestTransfersAddBalanceToDest(t *testing.T) {
	app, private := initAppTx(t)

	var initialDestNdau int64
	modifyDest(t, app, func(dest *backing.AccountData) {
		initialDestNdau = int64(dest.Balance)
	})

	const deltaNapu = 123 * constants.QuantaPerUnit

	tr := generateTransfer(t, 123, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifyDest(t, app, func(dest *backing.AccountData) {
		require.Equal(t, initialDestNdau+deltaNapu, int64(dest.Balance))
	})
}

func TestTransfersWhoseSrcAndDestAreEqualAreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	qty := int64(1)
	seq := uint64(1)

	// generate a transfer
	// this is almost a straight copy-paste of generateTransfer,
	// but we use source as dest as well
	//
	s, err := address.Validate(source)
	require.NoError(t, err)
	_, err = NewTransfer(
		s, s,
		math.Ndau(qty*constants.QuantaPerUnit),
		seq, []signature.PrivateKey{private},
	)
	require.Error(t, err)

	// We've just proved that this implementation refuses to generate
	// a transfer for which source and dest are identical.
	//
	// However, what if someone builds one from scratch?
	// We need to ensure that the application
	// layer rejects deserialized transfers which are invalid.
	tr := generateTransfer(t, qty, seq, []signature.PrivateKey{private})
	tr.Destination = tr.Source
	bytes := tr.SignableBytes()
	tr.Signatures = []signature.Signature{private.Sign(bytes)}

	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestInvalidTransactionDoesntAffectAnyBalance(t *testing.T) {
	app, private := initAppTx(t)
	var (
		beforeSrc  math.Ndau
		beforeDest math.Ndau
		afterSrc   math.Ndau
		afterDest  math.Ndau
	)
	modifySource(t, app, func(src *backing.AccountData) {
		beforeSrc = src.Balance
	})
	modifyDest(t, app, func(dest *backing.AccountData) {
		beforeDest = dest.Balance
	})

	// invalid: sequence 0
	tr := generateTransfer(t, 1, 0, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		afterSrc = src.Balance
	})
	modifyDest(t, app, func(dest *backing.AccountData) {
		afterDest = dest.Balance
	})

	require.Equal(t, beforeSrc, afterSrc)
	require.Equal(t, beforeDest, afterDest)
}

func TestTransfersOfMoreThanSourceBalanceAreInvalid(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(src *backing.AccountData) {
		src.Balance = 1 * constants.QuantaPerUnit
	})
	tr := generateTransfer(t, 2, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransfer(t, 1, 0, []signature.PrivateKey{private})
	resp := deliverTr(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransfer(t, 1, 1, []signature.PrivateKey{private})
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTransferWithExpiredEscrowsWorks(t *testing.T) {
	// setup app
	app, key, ts := initAppSettlement(t)
	require.True(t, app.blockTime.Compare(ts) >= 0)
	tn := ts.Add(1 * math.Second)

	// generate transfer
	// because the escrowed funds have cleared,
	// this should succeed
	s, err := address.Validate(settled)
	require.NoError(t, err)
	d, err := address.Validate(dest)
	require.NoError(t, err)
	tr, err := NewTransfer(
		s, d,
		math.Ndau(1),
		1, []signature.PrivateKey{key},
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTrAt(t, app, tr, tn)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTransferWithUnexpiredEscrowsFails(t *testing.T) {
	// setup app
	app, key, ts := initAppSettlement(t)
	// set app time to a day before the escrow expiry time
	tn := ts.Add(math.Duration(-24 * 3600 * math.Second))

	// generate transfer
	// because the escrowed funds have not yet cleared,
	// this should fail
	s, err := address.Validate(settled)
	require.NoError(t, err)
	d, err := address.Validate(dest)
	require.NoError(t, err)
	tr, err := NewTransfer(
		s, d,
		math.Ndau(1),
		1, []signature.PrivateKey{key},
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTrAt(t, app, tr, tn)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidationScriptValidatesTransfers(t *testing.T) {
	app, private := initAppTx(t)
	public2, private2, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// this script should be pretty stable for future versions of chaincode:
	// it means `one and not`, which just ensures that the first transfer key
	// is used, no matter how many keys are included
	script, err := base64.StdEncoding.DecodeString("oAAasUiI")
	require.NoError(t, err)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.ValidationScript = script
		ad.ValidationKeys = append(ad.ValidationKeys, public2)
	})

	t.Run("only first key", func(t *testing.T) {
		tr := generateTransfer(t, 123, 1, []signature.PrivateKey{private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 2, []signature.PrivateKey{private, private2})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := generateTransfer(t, 123, 3, []signature.PrivateKey{private2, private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := generateTransfer(t, 123, 4, []signature.PrivateKey{private2})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestTransferDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	dA, err := address.Validate(dest)
	require.NoError(t, err)

	for i := uint64(0); i < 2; i++ {
		modify(t, source, app, func(ad *backing.AccountData) {
			ad.Balance = math.Ndau(1 + i)
		})

		tx, err := NewTransfer(sA, dA, 1, 1+i, []signature.PrivateKey{private})
		require.NoError(t, err)

		resp := deliverTrWithTxFee(t, app, tx)

		var expect code.ReturnCode
		// this is different from the other TestXDeductsTxFee transactions:
		// we had to change the setup to account for the actual amount to be
		// transferred. The logic here is correct.
		if i > 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

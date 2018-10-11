package ndau

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/key"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
)

func generateRandomAddr(t *testing.T) string {
	seed, err := key.GenerateSeed(32)
	require.NoError(t, err)
	k, err := key.NewMaster(seed, key.NdauPrivateKeyID)
	require.NoError(t, err)
	a, err := address.Generate(address.Kind('a'), k.PubKeyBytes())
	return a.String()
}

func generateTransferAndLock(t *testing.T, destaddr string, qty int64, period math.Duration, seq uint64, keys []signature.PrivateKey) *TransferAndLock {
	s, err := address.Validate(source)
	require.NoError(t, err)
	if destaddr == "" {
		destaddr = dest
	}
	d, err := address.Validate(destaddr)
	require.NoError(t, err)
	tr, err := NewTransferAndLock(
		s, d,
		math.Ndau(qty*constants.QuantaPerUnit),
		period,
		seq, keys,
	)
	require.NoError(t, err)
	return tr
}

func TestTnLsWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []int64{0, -1, -2} {
		tr := generateTransferAndLock(t, generateRandomAddr(t), negQty, 999, uint64(idx+1), []signature.PrivateKey{private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestTnLsFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = backing.NewLock(90*math.Day, eai.DefaultLockBonusEAI)
	})

	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLsFromLockedButExpiredAddressesAreValid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		twoDaysAgo := now.Sub(math.Duration(2 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &twoDaysAgo
	})

	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestTnLsFromNotifiedAddressesAreInvalid(t *testing.T) {
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		tomorrow := now.Add(math.Duration(1 * math.Day))
		acct.Lock = backing.NewLock(1*math.Day, eai.DefaultLockBonusEAI)
		acct.Lock.UnlocksOn = &tomorrow
	})

	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLsDeductBalanceFromSource(t *testing.T) {
	app, private := initAppTx(t)

	var initialSourceNdau int64
	modifySource(t, app, func(src *backing.AccountData) {
		initialSourceNdau = int64(src.Balance)
	})

	const deltaNapu = 50 * constants.QuantaPerUnit

	tr := generateTransferAndLock(t, generateRandomAddr(t), 50, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		require.Equal(t, initialSourceNdau-deltaNapu, int64(src.Balance))
	})
}

func TestTnLsAddBalanceToDest(t *testing.T) {
	app, private := initAppTx(t)

	d := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, d, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, d, app, func(dest *backing.AccountData) {
		require.Equal(t, deltaNapu, int64(dest.Balance))
	})
}

func TestTnLsSetLockOnDest(t *testing.T) {
	app, private := initAppTx(t)

	d := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, d, 123, 90*math.Day, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, d, app, func(dest *backing.AccountData) {
		require.Equal(t, math.Duration(90*math.Day), dest.Lock.GetNoticePeriod())
	})
}

func TestTnLsSettlementPeriod(t *testing.T) {
	app, private := initAppTx(t)

	d := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	modifySource(t, app, func(src *backing.AccountData) {
		src.SettlementSettings.Period = 2 * math.Day
	})

	tr := generateTransferAndLock(t, d, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, d, app, func(dest *backing.AccountData) {
		s := []backing.Settlement{backing.Settlement{
			Qty:    123 * constants.QuantaPerUnit,
			Expiry: app.blockTime.Add(2 * math.Day),
		}}
		require.Equal(t, s, dest.Settlements)
	})
}

func TestTnLsFailForExistingDest(t *testing.T) {
	app, private := initAppTx(t)

	d := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, d, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, d, app, func(dest *backing.AccountData) {
		require.Equal(t, deltaNapu, int64(dest.Balance))
	})

	tr = generateTransferAndLock(t, d, 123, 888, 2, []signature.PrivateKey{private})
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLsWhoseSrcAndDestAreEqualAreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	qty := int64(1)
	seq := uint64(1)

	// generate a transfer
	// this is almost a straight copy-paste of generateTransferAndLock,
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
	tr := generateTransferAndLock(t, generateRandomAddr(t), qty, 888, seq, []signature.PrivateKey{private})
	tr.Destination = tr.Source
	bytes := tr.SignableBytes()
	tr.Signatures = []signature.Signature{private.Sign(bytes)}

	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestInvalidTnLTransactionDoesntAffectAnyBalance(t *testing.T) {
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
	tr := generateTransferAndLock(t, "", 1, 0, 999, []signature.PrivateKey{private})
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

func TestTnLsOfMoreThanSourceBalanceAreInvalid(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(src *backing.AccountData) {
		src.Balance = 1 * constants.QuantaPerUnit
	})
	tr := generateTransferAndLock(t, generateRandomAddr(t), 2, 888, 1, []signature.PrivateKey{private})
	resp := deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransferAndLock(t, generateRandomAddr(t), 1, 999, 0, []signature.PrivateKey{private})
	resp := deliverTr(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTr(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLWithExpiredEscrowsWorks(t *testing.T) {
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

func TestTnLWithUnexpiredEscrowsFails(t *testing.T) {
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

func TestValidationScriptValidatesTnLs(t *testing.T) {
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
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 888, 1, []signature.PrivateKey{private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 2, []signature.PrivateKey{private, private2})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 3, []signature.PrivateKey{private2, private})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 4, []signature.PrivateKey{private2})
		resp := deliverTr(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestTnLDeductsTxFee(t *testing.T) {
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

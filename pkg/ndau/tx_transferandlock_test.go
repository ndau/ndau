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
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func generateRandomAddr(t *testing.T) address.Address {
	seed, err := key.GenerateSeed(32)
	require.NoError(t, err)
	k, err := key.NewMaster(seed)
	require.NoError(t, err)
	a, err := address.Generate(address.KindUser, k.PubKeyBytes())
	require.NoError(t, err)
	return a
}

func generateTransferAndLock(t *testing.T, destaddr address.Address, qty int64, period math.Duration, seq uint64, keys []signature.PrivateKey) *TransferAndLock {
	tr := NewTransferAndLock(
		sourceAddress, destaddr,
		math.Ndau(qty*constants.QuantaPerUnit),
		period,
		seq, keys...,
	)
	return tr
}

func TestTnLsWhoseQtyLTE0AreInvalid(t *testing.T) {
	app, private := initAppTx(t)

	for idx, negQty := range []int64{0, -1, -2} {
		tr := generateTransferAndLock(t, generateRandomAddr(t), negQty, 999, uint64(idx+1), []signature.PrivateKey{private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	}
}

func TestTnLsFromLockedAddressesProhibited(t *testing.T) {
	app, private := initAppTx(t)
	modifySource(t, app, func(acct *backing.AccountData) {
		acct.Lock = backing.NewLock(90*math.Day, eai.DefaultLockBonusEAI)
	})

	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
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
	resp := deliverTx(t, app, tr)
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
	resp := deliverTx(t, app, tr)
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
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modifySource(t, app, func(src *backing.AccountData) {
		require.Equal(t, initialSourceNdau-deltaNapu, int64(src.Balance))
	})
}

func TestTnLsAddBalanceToDest(t *testing.T) {
	app, private := initAppTx(t)

	destAddress := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, destAddress, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, destAddress.String(), app, func(dest *backing.AccountData) {
		require.Equal(t, deltaNapu, int64(dest.Balance))
	})
}

func TestTnLsSetLockOnDest(t *testing.T) {
	app, private := initAppTx(t)

	destAddress := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, destAddress, 123, 90*math.Day, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, destAddress.String(), app, func(dest *backing.AccountData) {
		require.Equal(t, math.Duration(90*math.Day), dest.Lock.GetNoticePeriod())
	})
}

func TestTnLsSettlementPeriod(t *testing.T) {
	app, private := initAppTx(t)

	destAddress := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	modifySource(t, app, func(src *backing.AccountData) {
		src.RecourseSettings.Period = 2 * math.Day
	})

	tr := generateTransferAndLock(t, destAddress, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// can't require equality because we don't care about the tx hash etc
	dest, _ := app.getAccount(destAddress)
	require.Equal(t, len(dest.Holds), 1)
	require.Equal(t, math.Ndau(123*constants.QuantaPerUnit), dest.Holds[0].Qty)
	require.NotNil(t, dest.Holds[0].Expiry)
	require.Equal(t, app.BlockTime().Add(2*math.Day), *dest.Holds[0].Expiry)
}

func TestTnLsFailForExistingDest(t *testing.T) {
	app, private := initAppTx(t)

	destAddress := generateRandomAddr(t)
	const deltaNapu = int64(123 * constants.QuantaPerUnit)

	tr := generateTransferAndLock(t, destAddress, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	modify(t, destAddress.String(), app, func(dest *backing.AccountData) {
		require.Equal(t, deltaNapu, int64(dest.Balance))
	})

	tr = generateTransferAndLock(t, destAddress, 123, 888, 2, []signature.PrivateKey{private})
	resp = deliverTx(t, app, tr)
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
	tr := NewTransferAndLock(
		sourceAddress, sourceAddress,
		math.Ndau(qty*constants.QuantaPerUnit),
		math.Second,
		seq, private,
	)

	// We used to require that the constructors refuse to construct an invalid
	// transaction, but the philosophy has changed: constructors are generated
	// now, and are just shorthand for constructing an object manually and
	// optionally signing it. As such, we can't depend on any particular
	// validation in the constructor; we have to depend on the node to reject
	// invalid transactions. (This has the side benefit of keeping all validation
	// logic in exactly one place.)
	//
	// We need to ensure that the application
	// layer rejects deserialized transfers which are invalid.

	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLSignatureMustValidate(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	// I'm almost completely certain that this will be an invalid signature
	sig, err := signature.RawSignature(signature.Ed25519, make([]byte, signature.Ed25519.SignatureSize()))
	require.NoError(t, err)
	tr.Signatures = []signature.Signature{*sig}
	resp := deliverTx(t, app, tr)
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
	tr := generateTransferAndLock(t, destAddress, 1, 0, 999, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
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
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLSequenceMustIncrease(t *testing.T) {
	app, private := initAppTx(t)
	invalidZero := generateTransferAndLock(t, generateRandomAddr(t), 1, 999, 0, []signature.PrivateKey{private})
	resp := deliverTx(t, app, invalidZero)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	// valid now, because its sequence is greater than the account sequence
	tr := generateTransferAndLock(t, generateRandomAddr(t), 1, 888, 1, []signature.PrivateKey{private})
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	// invalid now because account sequence must have been updated
	resp = deliverTx(t, app, tr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestTnLWithExpiredEscrowsWorks(t *testing.T) {
	// setup app
	app, key, ts := initAppSettlement(t)
	require.True(t, app.BlockTime().Compare(ts) >= 0)
	tn := ts.Add(1 * math.Second)

	// generate transfer
	// because the escrowed funds have cleared,
	// this should succeed
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	destAddress := destAddress
	require.NoError(t, err)
	tr := NewTransferAndLock(
		sourceAddress, destAddress,
		math.Ndau(1),
		math.Second,
		1, key,
	)
	require.NoError(t, err)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
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
	sourceAddress, err := address.Validate(settled)
	require.NoError(t, err)
	destAddress := destAddress
	require.NoError(t, err)
	tr := NewTransferAndLock(
		sourceAddress, destAddress,
		math.Ndau(1),
		math.Second,
		1, key,
	)

	// send transfer
	resp := deliverTxAt(t, app, tr, tn)
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
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys in order", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 2, []signature.PrivateKey{private, private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("both keys out of order", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 3, []signature.PrivateKey{private2, private})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	})
	t.Run("only second key", func(t *testing.T) {
		tr := generateTransferAndLock(t, generateRandomAddr(t), 123, 999, 4, []signature.PrivateKey{private2})
		resp := deliverTx(t, app, tr)
		require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	})
}

func TestTnLDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	for i := uint64(0); i < 2; i++ {
		modify(t, source, app, func(ad *backing.AccountData) {
			ad.Balance = math.Ndau(1 + i)
		})

		tx := NewTransferAndLock(sourceAddress, destAddress, 1, math.Second, 1+i, private)
		resp := deliverTxWithTxFee(t, app, tx)

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

func TestTnLsPreventsClaimingExchangeAccount(t *testing.T) {
	app, private := initAppTx(t)

	destPublic, destPrivate, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	destAddress, err := address.Generate(address.KindUser, destPublic.KeyBytes())
	require.NoError(t, err)

	tr := generateTransferAndLock(
		t, destAddress, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// If the dest is an exchange address, we shouldn't be able to claim the locked account.
	context := ddc(t).withExchangeAccount(destAddress)

	ca := NewSetValidation(
		destAddress,
		destPublic,
		[]signature.PublicKey{newPublic},
		[]byte{},
		2,
		destPrivate,
	)

	// The ClaimAccount should fail.
	resp, _ = deliverTxContext(t, app, ca, context)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// The dest should still be locked after the TransferAndLock.
	destAcct, _ := app.getAccount(destAddress)
	require.True(t, destAcct.IsLocked(app.BlockTime()))
}

func TestTnLsPreventsClaimingExchangeAccountAsChild(t *testing.T) {
	app, private := initAppTx(t)

	tr := generateTransferAndLock(
		t, childAddress, 123, 888, 1, []signature.PrivateKey{private})
	resp := deliverTx(t, app, tr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	newPublic, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// If the source is an exchange address, we shouldn't be able to claim the locked child.
	context := ddc(t).withExchangeAccount(sourceAddress)

	cca := NewCreateChildAccount(
		sourceAddress,
		childAddress,
		childPublic,
		childSignature,
		childSettlementPeriod,
		[]signature.PublicKey{newPublic},
		[]byte{},
		childAddress,
		2,
		private,
	)

	resp, _ = deliverTxContext(t, app, cca, context)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// The child should still be locked after the TransferAndLock.
	childAcct, _ := app.getAccount(childAddress)
	require.True(t, childAcct.IsLocked(app.BlockTime()))
}

package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestValidLockTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	lock := NewLock(sourceAddress, math.Duration(30*math.Day), 1, private)
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestLockAccountValidates(t *testing.T) {
	app, private := initAppTx(t)
	lock := NewLock(sourceAddress, math.Duration(30*math.Day), 1, private)

	// make the account field invalid
	lock.Target = address.Address{}
	lock.Signatures = []signature.Signature{private.Sign(lock.SignableBytes())}

	// compute must be invalid
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestLockSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	lock := NewLock(sourceAddress, math.Duration(30*math.Day), 0, private)

	// lock must be invalid
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestLockSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	lock := NewLock(sourceAddress, math.Duration(30*math.Day), 1, private)

	// flip a single bit in the signature
	sigBytes := lock.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	lock.Signatures[0] = *wrongSignature

	// lock must be invalid
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestLockChangesAppState(t *testing.T) {
	duration := math.Duration(30 * math.Day)
	app, private := initAppTx(t)
	lock := NewLock(sourceAddress, duration, 1, private)

	acct, _ := app.getAccount(sourceAddress)
	require.Nil(t, acct.Lock)

	resp := deliverTx(t, app, lock)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	acct, _ = app.getAccount(sourceAddress)
	require.NotNil(t, acct.Lock)
	require.Equal(t, duration, acct.Lock.NoticePeriod)
	require.Nil(t, acct.Lock.UnlocksOn)
}

func TestLockCannotReduceLockLength(t *testing.T) {
	// set up fixture: source acct must already be locked
	duration := math.Duration(30 * math.Day)
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = backing.NewLock(duration, eai.DefaultLockBonusEAI)
	})

	// construct invalid relock tx
	lock := NewLock(sourceAddress, math.Duration(int64(duration)-1), 1, private)

	// lock must be invalid
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRelockNotified(t *testing.T) {
	// set up fixture: source acct must already be locked and notified
	duration := math.Duration(30 * math.Day)
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ts := math.Timestamp(int64(duration))
		ad.Lock = backing.NewLock(duration, eai.DefaultLockBonusEAI)
		ad.Lock.UnlocksOn = &ts
	})

	// construct relock tx of half original duration
	newDuration := math.Duration(int64(duration) / 2)
	lock := NewLock(sourceAddress, newDuration, 1, private)

	// lock must be invalid before halfway point of notice period
	halfway := math.Timestamp(int64(duration) / 2)
	resp := deliverTxAt(t, app, lock, halfway.Sub(math.Duration(1)))
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))

	// lock must be valid on and after halfway point of notice period
	lock = NewLock(sourceAddress, newDuration, 2, private)
	resp = deliverTxAt(t, app, lock, halfway)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// relock must have reset lock and cleared notification
	acct, _ := app.getAccount(sourceAddress)
	require.NotNil(t, acct.Lock)
	require.Equal(t, newDuration, acct.Lock.NoticePeriod)
	require.Nil(t, acct.Lock.UnlocksOn)
}

func TestLockDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewLock(sourceAddress, math.Duration(30*math.Day), 1+uint64(i), private)

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

func TestCannotLockExchangeAccount(t *testing.T) {
	app, private := initAppTx(t)

	app.setExchangeAccount(sourceAddress.String())

	lock := NewLock(sourceAddress, math.Duration(30*math.Day), 1, private)
	bytes, err := tx.Marshal(lock, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

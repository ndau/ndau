package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func initAppNotify(t *testing.T) (*App, signature.PrivateKey) {
	duration := math.Duration(30 * math.Day)
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = backing.NewLock(duration, eai.DefaultLockBonusEAI)
	})
	return app, private
}

func TestValidNotifyTxIsValid(t *testing.T) {
	app, private := initAppNotify(t)
	notify := NewNotify(sourceAddress, 1, private)
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestNotifyAccountValidates(t *testing.T) {
	app, private := initAppNotify(t)
	notify := NewNotify(sourceAddress, 1, private)

	// make the account field invalid
	notify.Target = address.Address{}
	notify.Signatures = []signature.Signature{private.Sign(notify.SignableBytes())}

	// compute must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifySequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	notify := NewNotify(sourceAddress, 0, private)

	// notify must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifySignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	notify := NewNotify(sourceAddress, 1, private)

	// flip a single bit in the signature
	sigBytes := notify.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	notify.Signatures[0] = *wrongSignature

	// notify must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifyChangesAppState(t *testing.T) {
	app, private := initAppNotify(t)
	notify := NewNotify(sourceAddress, 1, private)

	acct, _ := app.getAccount(sourceAddress)
	require.NotNil(t, acct.Lock)
	require.Nil(t, acct.Lock.UnlocksOn)

	resp := deliverTx(t, app, notify)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	acct, _ = app.getAccount(sourceAddress)
	require.NotNil(t, acct.Lock.UnlocksOn)
}

func TestNotifyDeductsTxFee(t *testing.T) {
	app, private := initAppNotify(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := uint64(0); i < 2; i++ {
		tx := NewNotify(sourceAddress, 1+i, private)

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

func TestNotifyProperlyEndsLock(t *testing.T) {
	// inspired by a Real Bug!
	// https://github.com/oneiro-ndev/exchanges/blob/master/samples/btcec-secp256k1/ndau-test.sh
	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	app, sourceKey := initAppTx(t)

	// lock the source account, but it should be expired
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = backing.NewLock(1, eai.DefaultLockBonusEAI)
		ad.Lock.UnlocksOn = &now
	})

	// deliver the transfer at the very moment the source should unlock
	tx := NewTransfer(sourceAddress, destAddress, 1*constants.NapuPerNdau, 1, sourceKey)
	resp := deliverTxAt(t, app, tx, now)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// a side effect of noting that the account wasn't locked anymore should be
	// clearing the lock from the account data
	acct, _ := app.getAccount(sourceAddress)
	require.Nil(t, acct.Lock)
}

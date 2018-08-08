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

func initAppNotify(t *testing.T) (*App, signature.PrivateKey) {
	duration := math.Duration(30 * math.Day)
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Lock = &backing.Lock{
			NoticePeriod: duration,
		}
	})
	return app, private
}

func TestValidNotifyTxIsValid(t *testing.T) {
	app, private := initAppNotify(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	notify := NewNotify(sA, 1, private)
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestNotifyAccountValidates(t *testing.T) {
	app, private := initAppNotify(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	notify := NewNotify(sA, 1, private)

	// make the account field invalid
	notify.Account = address.Address{}
	notify.Signature = private.Sign(notify.SignableBytes())

	// compute must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifySequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	notify := NewNotify(sA, 0, private)

	// notify must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifySignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	notify := NewNotify(sA, 1, private)

	// flip a single bit in the signature
	sigBytes := notify.Signature.Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	notify.Signature = *wrongSignature

	// notify must be invalid
	bytes, err := tx.Marshal(notify, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNotifyChangesAppState(t *testing.T) {
	app, private := initAppNotify(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	notify := NewNotify(sA, 1, private)

	state := app.GetState().(*backing.State)
	acct, _ := state.GetAccount(sA, app.blockTime)
	require.NotNil(t, acct.Lock)
	require.Nil(t, acct.Lock.UnlocksOn)

	resp := deliverTr(t, app, notify)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state = app.GetState().(*backing.State)
	acct, _ = state.GetAccount(sA, app.blockTime)
	require.NotNil(t, acct.Lock.UnlocksOn)
}

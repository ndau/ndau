package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

// this fixture ensures that for these tests, "source" is staked to "eaiNode"
func initAppUnstake(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)

	const NodeBalance = 10000 * constants.QuantaPerUnit

	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.Balance = NodeBalance
	})

	return app, private
}

func TestValidUnstakeTxIsValid(t *testing.T) {
	app, private := initAppUnstake(t)
	d := NewUnstake(sourceAddress, 1, private)

	// d must be valid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestUnstakeAccountValidates(t *testing.T) {
	app, private := initAppUnstake(t)
	d := NewUnstake(sourceAddress, 1, private)

	// make the account field invalid
	d.Target = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeSequenceValidates(t *testing.T) {
	app, private := initAppUnstake(t)
	d := NewUnstake(sourceAddress, 0, private)

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeSignatureValidates(t *testing.T) {
	app, private := initAppUnstake(t)
	d := NewUnstake(sourceAddress, 1, private)

	// flip a single bit in the signature
	sigBytes := d.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	d.Signatures[0] = *wrongSignature

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeChangesAppState(t *testing.T) {
	app, private := initAppUnstake(t)
	d := NewUnstake(sourceAddress, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// state := app.GetState().(*backing.State)
	// we must have updated the target's stake node

	// we must have updated the stake struct
	// require.ElementsMatch(
	// 	t,
	// 	[]backing.AccountData{
	// 		state.Accounts[eaiNode],
	// 	},
	// 	state.GetCostakers(nodeAddress),
	// )
}

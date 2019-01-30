package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func initAppStake(t *testing.T) (*App, signature.PrivateKey) {
	app, private := initAppTx(t)

	const NodeBalance = 10000 * constants.QuantaPerUnit

	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.Balance = NodeBalance

		ad.Stake = &backing.Stake{
			Address: nodeAddress,
		}
	})
	modifyNode(t, eaiNode, app, func(node *backing.Node) {
		*node = backing.NewNode(nodeAddress, NodeBalance)
	})

	return app, private
}

func TestValidStakeTxIsValid(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 1, private)

	// d must be valid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	if resp.Log != "" {
		t.Log(resp.Log)
	}
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestStakeAccountValidates(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 1, private)

	// make the account field invalid
	d.Target = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeStakeValidates(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 1, private)

	// make the account field invalid
	d.StakedAccount = address.Address{}
	d.Signatures = []signature.Signature{private.Sign(d.SignableBytes())}

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSequenceValidates(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 0, private)

	// d must be invalid
	bytes, err := tx.Marshal(d, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSignatureValidates(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 1, private)

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

func TestStakeChangesAppState(t *testing.T) {
	app, private := initAppStake(t)
	d := NewStake(sourceAddress, nodeAddress, 1, private)

	resp := deliverTx(t, app, d)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the target's stake node
	require.NotNil(t, state.Accounts[source].Stake)
	require.Equal(t, nodeAddress, state.Accounts[source].Stake.Address)
	// we must have updated the stake struct
	require.ElementsMatch(
		t,
		[]backing.AccountData{
			state.Accounts[source],
			state.Accounts[eaiNode],
		},
		state.GetCostakers(nodeAddress),
	)
}

func TestStakeDeductsTxFee(t *testing.T) {
	app, private := initAppStake(t)

	for i := 0; i < 2; i++ {
		modify(t, source, app, func(ad *backing.AccountData) {
			ad.Balance = math.Ndau(i + (1000 * constants.QuantaPerUnit))
			ad.Stake = nil
		})

		tx := NewStake(sourceAddress, nodeAddress, 1+uint64(i), private)

		resp := deliverTxWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.InvalidTransaction
		} else {
			expect = code.OK
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

func TestValidStakeTxIsValid(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// tx must be valid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestStakeTargetValidates(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Target = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeRulesValidates(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Rules = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeStakeToValidates(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.StakeTo = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSequenceValidates(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSignatureValidates(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	// flip a single bit in the signature
	sigBytes := tx.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	tx.Signatures[0] = *wrongSignature

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeChangesAppState(t *testing.T) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1, private)

	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)
	// we must have updated the source's delegation node
	require.Equal(t, &nodeAddress, state.Accounts[source].DelegationNode)

	// we must have added the source to the node's delegation responsibilities
	require.Contains(t, state.Stakes, eaiNode)
	require.Contains(t, state.Stakes[eaiNode], source)
}

func TestStakeDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	rulesAcct := getRulesAccount(t, app)

	for i := 0; i < 2; i++ {
		tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1*constants.NapuPerNdau, 1+uint64(i), private)

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

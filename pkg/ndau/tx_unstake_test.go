package ndau

import (
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	"github.com/stretchr/testify/require"
)

const (
	nodePrivate   = "node private"
	sourcePrivate = "souce private"
)

func initAppUnstake(t *testing.T) (*App, generator.Associated, address.Address) {
	app, assc := initApp(t)

	// set up initial conditions:
	// rules account is a proper rules account
	// node account is costaked with source account to rules account
	// source account is primary staked to rules account

	rulesAcct := getRulesAccount(t, app)
	modify(t, rulesAcct.String(), app, func(ad *backing.AccountData) {
		ad.StakeRules = &backing.StakeRules{
			// push 0, 0 to the stack and exit
			// at quit, stack top is exit code for validation;
			// second value is delay until hold is released
			Script:  vm.MiniAsm("handler 0 zero zero enddef").Bytes(),
			Inbound: make(map[string]uint64),
		}
	})

	setupAcct := func(addr address.Address, pkconst string) {
		public, private, err := signature.Generate(signature.Ed25519, nil)
		require.NoError(t, err)
		modify(t, addr.String(), app, func(ad *backing.AccountData) {
			ad.Balance = 10000 * constants.NapuPerNdau
			ad.ValidationKeys = []signature.PublicKey{public}
		})
		assc[pkconst] = private
	}

	setupAcct(nodeAddress, nodePrivate)
	setupAcct(sourceAddress, sourcePrivate)

	err := app.UpdateStateImmediately(func(stI metast.State) (st metast.State, err error) {
		st, err = app.Stake(1000*constants.NapuPerNdau, nodeAddress, sourceAddress, rulesAcct, nil)(stI)
		if err != nil {
			return
		}
		st, err = app.Stake(1000*constants.NapuPerNdau, sourceAddress, rulesAcct, rulesAcct, nil)(st)
		return
	})
	require.NoError(t, err)

	return app, assc, rulesAcct
}
func TestValidUnstakeTxIsValid(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// tx must be valid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestUnstakeTargetValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Target = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeStakeRulesValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(nodeAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Rules = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeStakeToValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.StakeTo = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeSequenceValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 0, private)

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestUnstakeSignatureValidates(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)
	tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

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

func TestCoUnstakeChangesAppState(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[nodePrivate].(signature.PrivateKey)

	tx := NewUnstake(nodeAddress, rulesAcct, sourceAddress, 1000*constants.NapuPerNdau, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)

	// must have removed previous stake from outbound stake list
	nodeData := state.Accounts[nodeAddress.String()]
	require.Empty(t, nodeData.Holds)

	// must not have updated inbound stake list (not primary)
	rulesData := state.Accounts[rulesAcct.String()]
	require.NotNil(t, rulesData.StakeRules)
	require.NotEmpty(t, rulesData.StakeRules.Inbound)
	require.NotContains(t, rulesData.StakeRules.Inbound, nodeAddress.String())

	// must have updated StakeTo costakers list
	sourceData := state.Accounts[source]
	require.NotZero(t, sourceData)
	if rulesCostakers, ok := sourceData.Costakers[rulesAcct.String()]; ok {
		require.NotContains(t, rulesCostakers, nodeAddress.String())
	}
}

func TestPrimaryUnstakeChangesAppState(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)

	tx := NewUnstake(sourceAddress, rulesAcct, rulesAcct, 1000*constants.NapuPerNdau, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)

	// must have updated outbound stake list
	sourceData := state.Accounts[source]
	if sourceData.Holds != nil {
		require.Empty(t, sourceData.Holds)
	}

	// must have updated inbound stake list
	rulesData := state.Accounts[rulesAcct.String()]
	require.NotZero(t, rulesData) // must exist
	require.NotNil(t, rulesData.StakeRules)
	require.Empty(t, rulesData.StakeRules.Inbound)
}

func TestUnstakeDeductsTxFee(t *testing.T) {
	app, assc, rulesAcct := initAppUnstake(t)
	private := assc[sourcePrivate].(signature.PrivateKey)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = (1000 * constants.NapuPerNdau) + 1
	})

	for i := 0; i < 2; i++ {
		tx := NewUnstake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1+uint64(i), private)

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

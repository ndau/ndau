package ndau

import (
	"testing"

	"github.com/oneiro-ndev/chaincode/pkg/vm"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func initAppStake(t *testing.T) (*App, signature.PrivateKey, address.Address) {
	app, private := initAppTx(t)
	rulesAcct := getRulesAccount(t, app)

	modify(t, rulesAcct.String(), app, func(ad *backing.AccountData) {
		ad.StakeRules = &backing.StakeRules{
			// push 0 to the stack and exit
			// at quit, stack top is exit code for validation
			Script:  vm.MiniAsm("handler 0 zero enddef").Bytes(),
			Inbound: make(map[string]uint64),
		}
	})

	nodeData, _ := app.getAccount(nodeAddress)
	modify(t, nodeAddress.String(), app, func(ad *backing.AccountData) {
		ad = &nodeData
	})

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 10000 * constants.NapuPerNdau
	})

	return app, private, rulesAcct
}
func TestValidStakeTxIsValid(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// tx must be valid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

}

func TestStakeTargetValidates(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Target = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeRulesValidates(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.Rules = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeStakeToValidates(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

	// make the account field invalid
	tx.StakeTo = address.Address{}
	tx.Signatures = []signature.Signature{private.Sign(tx.SignableBytes())}

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSequenceValidates(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 0, private)

	// tx must be invalid
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeSignatureValidates(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)
	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)

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

func TestCoStakeChangesAppState(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)

	tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)

	// must have updated outbound stake list
	sourceData := state.Accounts[source]
	require.NotNil(t, sourceData.Holds)
	require.Equal(t, len(sourceData.Holds), 1)
	require.Equal(t, math.Ndau(1000*constants.NapuPerNdau), sourceData.Holds[0].Qty)
	require.NotNil(t, sourceData.Holds[0].Stake)
	require.Equal(t, rulesAcct, sourceData.Holds[0].Stake.RulesAcct)
	require.Equal(t, nodeAddress, sourceData.Holds[0].Stake.StakeTo)

	// must not have updated inbound stake list (not primary)
	rulesData := state.Accounts[rulesAcct.String()]
	require.NotZero(t, rulesData) // must exist
	if rulesData.StakeRules != nil {
		require.Empty(t, rulesData.StakeRules.Inbound)
	}

	// must have updated StakeTo costakers list
	nodeData := state.Accounts[nodeAddress.String()]
	require.NotZero(t, nodeData)
	require.Contains(t, nodeData.Costakers, rulesAcct.String())
	require.NotNil(t, nodeData.Costakers[rulesAcct.String()])
	require.Contains(t, nodeData.Costakers[rulesAcct.String()], source)
}

func TestPrimaryStakeChangesAppState(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)

	tx := NewStake(sourceAddress, rulesAcct, rulesAcct, 1000*constants.NapuPerNdau, 1, private)
	resp := deliverTx(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state := app.GetState().(*backing.State)

	// must have updated outbound stake list
	sourceData := state.Accounts[source]
	require.NotNil(t, sourceData.Holds)
	require.Equal(t, len(sourceData.Holds), 1)
	require.Equal(t, math.Ndau(1000*constants.NapuPerNdau), sourceData.Holds[0].Qty)
	require.NotNil(t, sourceData.Holds[0].Stake)
	require.Equal(t, rulesAcct, sourceData.Holds[0].Stake.RulesAcct)
	require.Equal(t, rulesAcct, sourceData.Holds[0].Stake.StakeTo)

	// must have updated inbound stake list
	rulesData := state.Accounts[rulesAcct.String()]
	require.NotZero(t, rulesData) // must exist
	require.NotNil(t, rulesData.StakeRules)
	require.Contains(t, rulesData.StakeRules.Inbound, source)

	// must not be possible to double primary stake
	tx = NewStake(sourceAddress, rulesAcct, rulesAcct, 1000*constants.NapuPerNdau, 2, private)
	resp = deliverTx(t, app, tx)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestStakeDeductsTxFee(t *testing.T) {
	app, private, rulesAcct := initAppStake(t)

	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = (1000 * constants.NapuPerNdau) + 1
	})

	for i := 0; i < 2; i++ {
		tx := NewStake(sourceAddress, rulesAcct, nodeAddress, 1000*constants.NapuPerNdau, 1+uint64(i), private)

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

package ndau

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const cnrStake = 1000 * constants.QuantaPerUnit

func initAppCNR(t *testing.T) (*App, signature.PrivateKey, math.Timestamp) {
	app, _ := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// generate the transfer key so we claim the node reward
	public, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	modify(t, eaiNode, app, func(acct *backing.AccountData) {
		acct.TransferKeys = []signature.PublicKey{public}
	})

	// see chaincode/chasm/examples/distribution.chasm
	script, err := base64.StdEncoding.DecodeString("oAAOAlIFcD2QCYIACoiAAAJgPQ4CRiFQIWRGiA==")
	require.NoError(t, err)

	costakers := make(map[string]math.Ndau)
	totalStake := math.Ndau(0)
	for _, addr := range []string{eaiNode, source} {
		modify(t, addr, app, func(ad *backing.AccountData) {
			ad.Balance = cnrStake
		})
		costakers[addr] = cnrStake
		totalStake += cnrStake
	}

	now, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	eA, err := address.Validate(eaiNode)
	require.NoError(t, err)

	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		st.Nodes[eaiNode] = backing.Node{
			Active:             true,
			DistributionScript: script,
			Costakers:          costakers,
			TotalStake:         totalStake,
		}

		st.LastNodeRewardNomination = now
		st.UnclaimedNodeReward = 100 * constants.QuantaPerUnit
		// given the distribution script in place and this quantity of
		// ndau, we expect 60 ndau to be disbursed to eaiNode and
		// 40 ndau to source
		st.NodeRewardWinner = eA

		return st, nil
	})

	return app, private, now
}
func TestValidClaimNodeRewardTxIsValid(t *testing.T) {
	app, private, _ := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(nA, 1, []signature.PrivateKey{private})

	bytes, err := tx.Marshal(cnr, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	t.Log(resp.Log)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
}

func TestClaimNodeRewardAccountValidates(t *testing.T) {
	app, private, _ := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(nA, 1, []signature.PrivateKey{private})

	// make the account field invalid
	cnr.Node = address.Address{}
	cnr.Signatures = []signature.Signature{private.Sign(cnr.SignableBytes())}

	// must be invalid
	bytes, err := tx.Marshal(cnr, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimNodeRewardSequenceValidates(t *testing.T) {
	app, private, _ := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(nA, 0, []signature.PrivateKey{private})

	// cnr must be invalid
	bytes, err := tx.Marshal(cnr, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimNodeRewardSignatureValidates(t *testing.T) {
	app, private, _ := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(nA, 1, []signature.PrivateKey{private})

	// flip a single bit in the signature
	sigBytes := cnr.Signatures[0].Bytes()
	sigBytes[0] = sigBytes[0] ^ 1
	wrongSignature, err := signature.RawSignature(signature.Ed25519, sigBytes)
	require.NoError(t, err)
	cnr.Signatures[0] = *wrongSignature

	// cnr must be invalid
	bytes, err := tx.Marshal(cnr, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestClaimNodeRewardChangesAppState(t *testing.T) {
	app, private, now := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(nA, 1, []signature.PrivateKey{private})

	state := app.GetState().(*backing.State)
	acct, _ := state.GetAccount(nA, app.blockTime)
	require.Equal(t, math.Ndau(cnrStake), acct.Balance)

	resp := deliverTrAt(t, app, cnr, now+1)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	state = app.GetState().(*backing.State)
	acct, _ = state.GetAccount(nA, app.blockTime)

	// we expect 60% of node rewards to go to the EAI node in this setup
	require.Equal(t, math.Ndau(cnrStake+(60*constants.QuantaPerUnit)), acct.Balance)

	// we expect 40% of node rewards to go to source acct in this setup
	sA, err := address.Validate(source)
	require.NoError(t, err)
	acct, _ = state.GetAccount(sA, app.blockTime)
	require.Equal(t, math.Ndau(cnrStake+(40*constants.QuantaPerUnit)), acct.Balance)
}

func TestClaimNodeRewardDeductsTxFee(t *testing.T) {
	app, private, _ := initAppCNR(t)
	nA, err := address.Validate(eaiNode)
	require.NoError(t, err)

	modify(t, eaiNode, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	for i := 0; i < 2; i++ {
		tx := NewClaimNodeReward(nA, 1+uint64(i), []signature.PrivateKey{private})

		resp := deliverTrWithTxFee(t, app, tx)

		var expect code.ReturnCode
		if i == 0 {
			expect = code.OK
		} else {
			expect = code.InvalidTransaction
		}
		require.Equal(t, expect, code.ReturnCode(resp.Code))
	}
}

func TestImpostorCannotClaimNodeReward(t *testing.T) {
	app, private, _ := initAppCNR(t)
	sA, err := address.Validate(source)
	require.NoError(t, err)
	cnr := NewClaimNodeReward(sA, 1, []signature.PrivateKey{private})

	bytes, err := tx.Marshal(cnr, TxIDs)
	require.NoError(t, err)
	resp := app.CheckTx(bytes)
	t.Log(resp.Log)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

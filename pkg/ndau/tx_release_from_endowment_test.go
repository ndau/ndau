package ndau

import (
	"fmt"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const rfeKeys = "rfe private keys"

func initAppRFE(t *testing.T) (*App, config.MockAssociated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	// fetch the RFE address system variable
	rfeAddr := address.Address{}
	err := app.System(sv.ReleaseFromEndowmentAddressName, &rfeAddr)
	require.NoError(t, err)
	assc[rfeKeys], err = MockSystemAccount(app, rfeAddr)

	return app, assc
}

func TestRFEIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				1,
				[]signature.PrivateKey{private},
			)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestRFEIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRFE(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)

	rfe := NewReleaseFromEndowment(
		math.Ndau(1),
		targetAddress,
		1,
		[]signature.PrivateKey{private},
	)

	rfeBytes, err := tx.Marshal(&rfe, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(rfeBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRFEAddsNdauToExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				ad.Balance = math.Ndau(10)
			})

			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				uint64(i+1),
				[]signature.PrivateKey{private},
			)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
				Time: time.Now().Unix(),
			}})
			dresp := app.DeliverTx(rfeBytes)
			app.EndBlock(abci.RequestEndBlock{})
			app.Commit()

			require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				require.Equal(t, math.Ndau(11), ad.Balance)
			})
		})
	}
}

func TestValidRFEAddsNdauToNonExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)

			targetAddress, err := address.Generate(address.KindUser, public.Bytes())
			require.NoError(t, err)

			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				uint64(i+1),
				[]signature.PrivateKey{private},
			)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
				Time: time.Now().Unix(),
			}})
			dresp := app.DeliverTx(rfeBytes)
			app.EndBlock(abci.RequestEndBlock{})
			app.Commit()

			require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				require.Equal(t, math.Ndau(1), ad.Balance)
			})
		})
	}
}

func TestRFEIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[rfeKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.ReleaseFromEndowmentAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 rfe keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				uint64(i)+1,
				[]signature.PrivateKey{private},
			)

			resp := deliverTrWithTxFee(t, app, &rfe)

			var expect code.ReturnCode
			if i == 0 {
				expect = code.OK
			} else {
				expect = code.InvalidTransaction
			}
			require.Equal(t, expect, code.ReturnCode(resp.Code))
		})
	}
}
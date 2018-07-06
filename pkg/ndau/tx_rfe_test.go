package ndau

import (
	"fmt"
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndaunode/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
)

func initAppRFE(t *testing.T) (*App, config.MockAssociated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})
	return app, assc
}

func TestRFEIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			rfe, err := NewReleaseFromEndowment(math.Ndau(1), targetAddress, private)
			require.NoError(t, err)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestRFEIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRFE(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)

	rfe, err := NewReleaseFromEndowment(math.Ndau(1), targetAddress, private)
	require.NoError(t, err)

	rfeBytes, err := tx.Marshal(&rfe, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(rfeBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRFEAddsNdauToExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				ad.Balance = math.Ndau(10)
			})

			rfe, err := NewReleaseFromEndowment(math.Ndau(1), targetAddress, private)
			require.NoError(t, err)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
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
	privateKeys := assc[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)

			targetAddress, err := address.Generate(address.KindUser, public.Bytes())
			require.NoError(t, err)

			rfe, err := NewReleaseFromEndowment(math.Ndau(1), targetAddress, private)
			require.NoError(t, err)

			rfeBytes, err := tx.Marshal(&rfe, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(rfeBytes)
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

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

const rfeAddrs = "RFE addresses"

func initAppRFE(t *testing.T) (*App, config.MockAssociated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	rfePublic := make(sv.ReleaseFromEndowmentKeys, 0)
	err := app.System(sv.ReleaseFromEndowmentKeysName, &rfePublic)
	require.NoError(t, err)

	addresses := make([]address.Address, 0, len(rfePublic))
	for _, public := range rfePublic {
		addr, err := address.Generate(address.KindEndowment, public.Bytes())
		require.NoError(t, err)
		addresses = append(addresses, addr)
		// ensure that this address is tied to a real account
		modify(t, addr.String(), app, func(acct *backing.AccountData) {
			copy := public // make a copy so pointers are independent
			acct.TransferKeys = []signature.PublicKey{copy}
		})
	}
	assc[rfeAddrs] = addresses

	return app, assc
}

func TestRFEIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		txFeeAddr := assc[rfeAddrs].([]address.Address)[i]
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				txFeeAddr,
				1,
				private,
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
	app, assc := initAppRFE(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	txFeeAddr := assc[rfeAddrs].([]address.Address)[0]

	rfe := NewReleaseFromEndowment(
		math.Ndau(1),
		targetAddress,
		txFeeAddr,
		1,
		private,
	)

	rfeBytes, err := tx.Marshal(&rfe, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(rfeBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestValidRFEAddsNdauToExistingDestination(t *testing.T) {
	app, assc := initAppRFE(t)
	privateKeys := assc[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		txFeeAddr := assc[rfeAddrs].([]address.Address)[i]
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			modify(t, targetAddress.String(), app, func(ad *backing.AccountData) {
				ad.Balance = math.Ndau(10)
			})

			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				txFeeAddr,
				1,
				private,
			)

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
		txFeeAddr := assc[rfeAddrs].([]address.Address)[i]
		private := privateKeys[i]
		t.Run(fmt.Sprintf("private key: %x", private.Bytes()), func(t *testing.T) {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)

			targetAddress, err := address.Generate(address.KindUser, public.Bytes())
			require.NoError(t, err)

			rfe := NewReleaseFromEndowment(
				math.Ndau(1),
				targetAddress,
				txFeeAddr,
				1,
				private,
			)

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

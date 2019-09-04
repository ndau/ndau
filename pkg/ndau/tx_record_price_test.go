package ndau

import (
	"fmt"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/pricecurve"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const recordPriceKeys = "recordPrice private keys"

func initAppRecordPrice(t *testing.T) (*App, generator.Associated) {
	return initAppRecordPriceWithIndex(t, "", -1)
}

func initAppRecordPriceWithIndex(t *testing.T, indexAddr string, indexVersion int) (
	*App, generator.Associated,
) {
	app, assc := initAppWithIndex(t, indexAddr, indexVersion)
	app.InitChain(abci.RequestInitChain{})

	// fetch the RecordPrice address system variable
	recordPriceAddr := address.Address{}
	err := app.System(sv.RecordPriceAddressName, &recordPriceAddr)
	require.NoError(t, err)
	assc[recordPriceKeys], err = MockSystemAccount(app, recordPriceAddr)

	// ensure special acct contains exactly 1 napu so balance test works
	modify(t, recordPriceAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	return app, assc
}

func TestRecordPriceIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppRecordPrice(t)
	privateKeys := assc[recordPriceKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			recordPrice := NewRecordPrice(
				pricecurve.Nanocent(1),
				1,
				private,
			)

			recordPriceBytes, err := tx.Marshal(recordPrice, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: recordPriceBytes})
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestRecordPriceIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppRecordPrice(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	recordPrice := NewRecordPrice(
		pricecurve.Nanocent(1),
		1,
		private,
	)

	recordPriceBytes, err := tx.Marshal(recordPrice, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: recordPriceBytes})
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestRecordPriceIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppRecordPrice(t)
	privateKeys := assc[recordPriceKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.ReleaseFromEndowmentAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 recordPrice keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			recordPrice := NewRecordPrice(
				pricecurve.Nanocent(1),
				uint64(i)+1,
				private,
			)

			resp := deliverTxWithTxFee(t, app, recordPrice)

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

func TestMarketPriceMustBePositive(t *testing.T) {
	app, assc := initAppRecordPrice(t)
	privateKeys := assc[recordPriceKeys].([]signature.PrivateKey)

	recordPrice := NewRecordPrice(
		pricecurve.Nanocent(-1),
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, recordPrice)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
	require.Contains(t, resp.Log, "RecordPrice market price may not be <= 0")
}

func TestZeroMarketPriceMustIncurSIB(t *testing.T) {
	app, assc := initAppRecordPrice(t)
	privateKeys := assc[recordPriceKeys].([]signature.PrivateKey)

	recordPrice := NewRecordPrice(
		pricecurve.Nanocent(1),
		1,
		privateKeys...,
	)

	resp := deliverTx(t, app, recordPrice)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	require.NotZero(t, app.GetState().(*backing.State).SIB)

	t.Run("app query", func(t *testing.T) {
		resp := app.Query(abci.RequestQuery{
			Path: query.SIBEndpoint,
		})

		require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		require.NotEmpty(t, resp.Value)
		require.NotEmpty(t, resp.Info) // human-readable representation of value

		var sib query.SIBResponse
		leftovers, err := sib.UnmarshalMsg(resp.Value)
		require.NoError(t, err)
		require.Empty(t, leftovers)

		require.Equal(t, app.GetState().(*backing.State).SIB, sib.SIB)
	})
}

package ndau

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	tx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const nnrKeys = "nnr private keys"

func initAppNNR(t *testing.T) (*App, generator.Associated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	const qtyNodes = 2

	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.LastNodeRewardNomination = math.Timestamp(0)

		for i := 0; i < qtyNodes; i++ {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)
			addr, err := address.Generate(address.KindNdau, public.KeyBytes())
			require.NoError(t, err)

			state.Nodes[addr.String()] = backing.Node{
				TotalStake: math.Ndau(i + 1),
				Costakers:  make(map[string]math.Ndau),
				Active:     true,
			}

			state.Accounts[addr.String()] = backing.AccountData{
				Balance: math.Ndau(i + 1),
				Stake: &backing.Stake{
					Address: addr,
					Point:   math.Timestamp(0),
				},
			}
		}

		return state, nil
	})
	var err error
	app.blockTime, err = math.TimestampFrom(time.Now())
	require.NoError(t, err)

	// fetch the NNR address system variable
	nnrAddr := address.Address{}
	err = app.System(sv.NominateNodeRewardAddressName, &nnrAddr)
	require.NoError(t, err)
	assc[nnrKeys], err = MockSystemAccount(app, nnrAddr)

	return app, assc
}

func TestNNRIsValidWithValidSignature(t *testing.T) {
	app, assc := initAppNNR(t)
	privateKeys := assc[nnrKeys].([]signature.PrivateKey)

	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			nnr := NewNominateNodeReward(
				0,
				1,
				[]signature.PrivateKey{private},
			)

			nnrBytes, err := tx.Marshal(&nnr, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(nnrBytes)
			if resp.Log != "" {
				t.Log(resp.Log)
			}
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
		})
	}
}

func TestNNRIsInvalidWithInvalidSignature(t *testing.T) {
	app, _ := initAppNNR(t)
	_, private, err := signature.Generate(signature.Ed25519, nil)

	nnr := NewNominateNodeReward(
		0,
		1,
		[]signature.PrivateKey{private},
	)

	nnrBytes, err := tx.Marshal(&nnr, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(nnrBytes)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNNRIsValidOnlyWithSufficientTxFee(t *testing.T) {
	app, assc := initAppNNR(t)
	privateKeys := assc[nnrKeys].([]signature.PrivateKey)

	txFeeAddr := address.Address{}
	err := app.System(sv.NominateNodeRewardAddressName, &txFeeAddr)
	require.NoError(t, err)

	// with a tx fee of 1, only the first tx should succeed
	modify(t, txFeeAddr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	// our fixtures are set up with 2 nnr keys
	for i := 0; i < len(privateKeys); i++ {
		private := privateKeys[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			nnr := NewNominateNodeReward(
				0,
				uint64(i)+1,
				[]signature.PrivateKey{private},
			)

			resp := deliverTxWithTxFee(t, app, &nnr)

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

func TestNNRRequiresCooldown(t *testing.T) {
	app, assc := initAppNNR(t)
	privateKeys := assc[nnrKeys].([]signature.PrivateKey)

	nnr := NewNominateNodeReward(
		0,
		1,
		[]signature.PrivateKey{privateKeys[0]},
	)
	resp := deliverTx(t, app, &nnr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	nnr = NewNominateNodeReward(
		0, 2, []signature.PrivateKey{privateKeys[0]},
	)
	resp = deliverTx(t, app, &nnr)
	require.Equal(t, code.InvalidTransaction, code.ReturnCode(resp.Code))
}

func TestNNRCallsWebhook(t *testing.T) {
	// set up server listening on localhost
	qtyCalls := 0
	const port = ":31416"
	server := &http.Server{Addr: port}
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		qtyCalls++
		w.WriteHeader(204)
	})
	listener, err := net.Listen("tcp", port)
	require.NoError(t, err)
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Log(err)
		}
	}()
	defer server.Shutdown(context.Background())

	// edit app configuration to set webhook address
	app, assc := initAppNNR(t)
	webhookAddr := fmt.Sprintf("http://localhost%s", port)
	app.config.NodeRewardWebhook = &webhookAddr

	// now deliver the NNR transaction
	privateKeys := assc[nnrKeys].([]signature.PrivateKey)

	nnr := NewNominateNodeReward(
		0,
		1,
		[]signature.PrivateKey{privateKeys[0]},
	)
	resp := deliverTx(t, app, &nnr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	require.Equal(t, 1, qtyCalls)
}

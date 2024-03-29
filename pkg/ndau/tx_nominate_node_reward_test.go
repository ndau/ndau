package ndau

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	tx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	generator "github.com/ndau/system_vars/pkg/genesis.generator"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

const nnrKeys = "nnr private keys"

func initAppNNR(t *testing.T) (*App, generator.Associated) {
	app, assc := initApp(t)
	app.InitChain(abci.RequestInitChain{})

	rulesAcct, _ := getRulesAccount(t, app)

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
				Active: true,
			}

			state.Accounts[addr.String()] = backing.AccountData{
				Balance: math.Ndau(i + 1),
			}

			// self-stake this node
			stI, err = app.Stake(math.Ndau(i+1), addr, rulesAcct, rulesAcct, nil)(state)
			state = stI.(*backing.State)
			require.NoError(t, err)
		}

		return state, nil
	})

	// fetch the NNR address system variable
	nnrAddr := address.Address{}
	err := app.System(sv.NominateNodeRewardAddressName, &nnrAddr)
	require.NoError(t, err)
	assc[nnrKeys], err = MockSystemAccount(app, nnrAddr)
	require.NoError(t, err)

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
				private,
			)

			nnrBytes, err := tx.Marshal(nnr, TxIDs)
			require.NoError(t, err)

			resp := app.CheckTx(abci.RequestCheckTx{Tx: nnrBytes})
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
	require.NoError(t, err)

	nnr := NewNominateNodeReward(
		0,
		1,
		private,
	)

	nnrBytes, err := tx.Marshal(nnr, TxIDs)
	require.NoError(t, err)

	resp := app.CheckTx(abci.RequestCheckTx{Tx: nnrBytes})
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
				private,
			)

			resp := deliverTxWithTxFee(t, app, nnr)

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
		privateKeys[0],
	)
	resp := deliverTx(t, app, nnr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	nnr = NewNominateNodeReward(0, 2, privateKeys[0])
	resp = deliverTx(t, app, nnr)
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

	// set up a waitgroup so we can wait for the webhook to complete
	var wg sync.WaitGroup
	wg.Add(1)
	whDone = func() {
		wg.Done()
	}

	// now deliver the NNR transaction
	privateKeys := assc[nnrKeys].([]signature.PrivateKey)

	nnr := NewNominateNodeReward(
		0,
		1,
		privateKeys[0],
	)
	resp := deliverTx(t, app, nnr)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	wg.Wait()

	require.Equal(t, 1, qtyCalls)
}

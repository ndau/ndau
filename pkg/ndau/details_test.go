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
	"testing"

	"github.com/ndau/metanode/pkg/meta/app/code"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestTransactionsProduceUncreditedEAI(t *testing.T) {
	pub, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)

	// this is not an exhaustive list of transactions, but it should be sufficient
	// to demonstrate that uncredited EAI is reliably calculated and stored
	txs := []interface {
		NTransactable
		Signable
	}{
		NewTransfer(sourceAddress, destAddress, 1, 1),
		NewChangeValidation(sourceAddress, []signature.PublicKey{pub}, nil, 1),
		NewChangeRecoursePeriod(sourceAddress, 30*math.Day, 1),
		NewLock(sourceAddress, 90*math.Day, 1),
		NewSetRewardsDestination(sourceAddress, targetAddress, 1),
		NewTransferAndLock(sourceAddress, destAddress, 1, 90*math.Day, 1),
	}

	for _, tx := range txs {
		t.Run(metatx.NameOf(tx), func(t *testing.T) {
			app, _ := initApp(t)
			pub, pvt, err := signature.Generate(signature.Ed25519, nil)
			require.NoError(t, err)
			modify(t, source, app, func(ad *backing.AccountData) {
				ad.ValidationKeys = []signature.PublicKey{pub}
				ad.Balance = 1500 * constants.NapuPerNdau
			})
			sig := metatx.Sign(tx, pvt)
			tx.ExtendSignatures([]signature.Signature{sig})

			ad, _ := app.getAccount(sourceAddress)
			require.Zero(t, ad.UncreditedEAI)

			resp := deliverTxAt(t, app, tx, 45*math.Day)
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))

			ad, _ = app.getAccount(sourceAddress)
			require.NotZero(t, ad.UncreditedEAI)
		})
	}
}

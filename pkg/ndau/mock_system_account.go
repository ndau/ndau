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
	metast "github.com/ndau/metanode/pkg/meta/state"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
)

// MockSystemAccount generates a single mock system account
//
// given an address and the current time, generates some keypairs for this address,
// sets them in the app, and returns the private components
//
// mock accounts have 2 validation keys and no validation script; they
// therefore implement 1 of 2 multisig.
func MockSystemAccount(app *App, addr address.Address) ([]signature.PrivateKey, error) {
	const numKeys = 2

	publics := make([]signature.PublicKey, numKeys)
	privates := make([]signature.PrivateKey, numKeys)

	var err error
	for i := 0; i < numKeys; i++ {
		publics[i], privates[i], err = signature.Generate(signature.Ed25519, nil)
		if err != nil {
			return nil, err
		}
	}

	return privates, app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		st := stI.(*backing.State)

		acct, _ := app.getAccount(addr)
		acct.ValidationKeys = publics

		st.Accounts[addr.String()] = acct
		return st, nil
	})
}

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
	"time"

	"github.com/ndau/metanode/pkg/meta/app/code"
	"github.com/ndau/ndau/pkg/ndau/backing"
	math "github.com/ndau/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestCSPStoresPendingRecourseChange(t *testing.T) {
	app, private := initAppTx(t)
	acct := app.GetState().(*backing.State).Accounts[source]
	require.Equal(t, math.Duration(0), acct.RecourseSettings.Period)

	const newDuration = math.Duration(1 * math.Day)

	addr := sourceAddress

	cep := NewChangeRecoursePeriod(addr, newDuration, acct.Sequence+1, private)

	ts, err := math.TimestampFrom(time.Now())
	require.NoError(t, err)

	resp := deliverTxAt(t, app, cep, ts)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// the state of UpdateBalance is formally undefined at this point:
	// it's legit for implementations to notice that the current UpdateDuration
	// is 0, and to update the recourse period immediately. It's also legit
	// for them to wait for an UpdateRecourse call.

	// update the acct struct
	acct = app.GetState().(*backing.State).Accounts[source]
	acct.UpdateRecourses(ts)

	require.Equal(t, newDuration, acct.RecourseSettings.Period)
	require.Nil(t, acct.RecourseSettings.Next)
	require.Nil(t, acct.RecourseSettings.ChangesAt)
}

func TestChangeRecoursePeriodDeductsTxFee(t *testing.T) {
	app, private := initAppTx(t)
	modify(t, source, app, func(ad *backing.AccountData) {
		ad.Balance = 1
	})

	addr := sourceAddress

	const newDuration = math.Duration(1 * math.Day)

	for i := 0; i < 2; i++ {
		tx := NewChangeRecoursePeriod(
			addr,
			newDuration,
			uint64(i)+1,
			private,
		)

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

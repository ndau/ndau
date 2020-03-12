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
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/stretchr/testify/require"
)

func TestTxFeesAreAddedToPendingNodeRewards(t *testing.T) {
	app, private := initAppTx(t)

	pnr := app.GetState().(*backing.State).PendingNodeReward

	tx := NewTransfer(sourceAddress, destAddress, 1*constants.NapuPerNdau, 1, private)
	resp := deliverTxWithTxFee(t, app, tx)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	newPNR := app.GetState().(*backing.State).PendingNodeReward
	require.Equal(t, pnr+1, newPNR, "testing with tx fee of 1 should add 1 to PNR after single tx")
}

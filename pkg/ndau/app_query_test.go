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
	"fmt"
	"testing"

	"github.com/ndau/chaincode/pkg/vm"
	"github.com/ndau/metanode/pkg/meta/app/code"
	metast "github.com/ndau/metanode/pkg/meta/state"
	metatx "github.com/ndau/metanode/pkg/meta/transaction"
	"github.com/ndau/msgp-well-known-types/wkt"
	"github.com/ndau/ndau/pkg/ndau/backing"
	"github.com/ndau/ndau/pkg/query"
	"github.com/ndau/ndau/pkg/version"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/constants"
	"github.com/ndau/ndaumath/pkg/signature"
	math "github.com/ndau/ndaumath/pkg/types"
	sv "github.com/ndau/system_vars/pkg/system_vars"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestCanQueryAccountStatusSource(t *testing.T) {
	app, _ := initAppTx(t)

	// test the source, which should exist
	resp := app.Query(abci.RequestQuery{
		Path: query.AccountEndpoint,
		Data: []byte(source),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "acct exists: true", resp.Info)
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, math.Ndau(10000*constants.QuantaPerUnit), accountData.Balance)
}

func TestCanQueryAccountStatusDest(t *testing.T) {
	app, _ := initAppTx(t)

	// test the source, which should not exist
	resp := app.Query(abci.RequestQuery{
		Path: query.AccountEndpoint,
		Data: []byte(dest),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "acct exists: false", resp.Info)
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, math.Ndau(0), accountData.Balance)
}

func TestQueryRunsUpdateBalance(t *testing.T) {
	app, _, ts := initAppRecourse(t)
	t.Log("timestamp of end of recourse period", ts)
	t.Log("app blocktime", app.BlockTime())
	t.Log("comparison", app.BlockTime().Compare(ts))
	require.True(t, app.BlockTime().Compare(ts) >= 0)

	resp := app.Query(abci.RequestQuery{
		Path: query.AccountEndpoint,
		Data: []byte(settled),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.NotEqual(t, math.Ndau(0), accountData.Balance)
	require.Equal(t, 0, len(accountData.Holds))
}

func TestCanQuerySummary1(t *testing.T) {
	app, _ := initAppTx(t)

	resp := app.Query(abci.RequestQuery{
		Path: query.SummaryEndpoint,
		Data: nil,
	})
	expectedTotal := 10000 * constants.QuantaPerUnit
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, fmt.Sprintf("total ndau at height 1 is %d, in 1 accounts", expectedTotal), resp.Log)
	summary := new(query.Summary)
	_, err := summary.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.BlockHeight)
	require.Equal(t, 1, summary.NumAccounts)
	require.Equal(t, math.Ndau(expectedTotal), summary.TotalNdau)
}

func TestCanQuerySummary2(t *testing.T) {
	app, _ := initAppTx(t)
	// create a new account for different results
	public, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindUser, public.KeyBytes())
	require.NoError(t, err)
	modify(t, addr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1 * constants.QuantaPerUnit
	})
	app.SetHeight(2)

	resp := app.Query(abci.RequestQuery{
		Path: query.SummaryEndpoint,
		Data: nil,
	})
	expectedTotal := 10001 * constants.QuantaPerUnit
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, fmt.Sprintf("total ndau at height 2 is %d, in 2 accounts", expectedTotal), resp.Log)
	summary := new(query.Summary)
	_, err = summary.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, uint64(2), summary.BlockHeight)
	require.Equal(t, 2, summary.NumAccounts)
	require.Equal(t, math.Ndau(expectedTotal), summary.TotalNdau)
}

func TestCanQueryVersion(t *testing.T) {
	// this test can't pass unless you run it with ldflags set to inject
	// the version information properly. It exists mainly as an example
	// of how to use this query
	_, err := version.Get()
	if err != nil {
		t.Skip("version not set by linker. See `go build -extldflags`")
	}

	app, _ := initApp(t)

	resp := app.Query(abci.RequestQuery{
		Path: query.VersionEndpoint,
	})

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.NotEmpty(t, resp.Value)
}

func TestPrevalidateValidTx(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	trb, err := metatx.Marshal(tr, TxIDs)
	require.NoError(t, err)

	resp := app.Query(abci.RequestQuery{
		Path: query.PrevalidateEndpoint,
		Data: trb,
	})

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.NotEmpty(t, resp.Info)

	var fee, sib math.Ndau
	_, err = fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
}

func TestPrevalidateInvalidTx(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 50, 0, []signature.PrivateKey{private})
	trb, err := metatx.Marshal(tr, TxIDs)
	require.NoError(t, err)

	resp := app.Query(abci.RequestQuery{
		Path: query.PrevalidateEndpoint,
		Data: trb,
	})

	require.Equal(t, code.QueryError, code.ReturnCode(resp.Code))
	require.NotEmpty(t, resp.Info)

	var fee, sib math.Ndau
	_, err = fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
}

func TestPrevalidateReportsCorrectFee(t *testing.T) {
	app, private := initAppTx(t)
	tr := generateTransfer(t, 50, 1, []signature.PrivateKey{private})
	trb, err := metatx.Marshal(tr, TxIDs)
	require.NoError(t, err)

	// set up a delivery context where a tx has a known fee
	dc := ddc(t).with(func(sysvars map[string][]byte) {
		script := vm.MiniAsm("handler 0 one enddef").Bytes()
		msgp, err := wkt.Bytes(script).MarshalMsg(nil)
		require.NoError(t, err)
		sysvars[sv.TxFeeScriptName] = msgp
	})
	var resp abci.ResponseQuery
	dc.Within(app, func() {
		resp = app.Query(abci.RequestQuery{
			Path: query.PrevalidateEndpoint,
			Data: trb,
		})
	})

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.NotEmpty(t, resp.Info)

	var fee, sib math.Ndau
	_, err = fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
	require.Equal(t, math.Ndau(1), fee)
}

func TestPrevalidateReportsCorrectSIB(t *testing.T) {
	app, private := initAppTx(t)
	tr := NewTransfer(sourceAddress, destAddress, 2*constants.NapuPerNdau, 1, private)
	trb, err := metatx.Marshal(tr, TxIDs)
	require.NoError(t, err)

	// set 50% SIB
	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.SIB = constants.RateDenominator / 2
		return state, nil
	})

	resp := app.Query(abci.RequestQuery{
		Path: query.PrevalidateEndpoint,
		Data: trb,
	})

	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.NotEmpty(t, resp.Info)

	var fee, sib math.Ndau
	_, err = fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
	require.Equal(t, math.Ndau(constants.NapuPerNdau), sib)
}

func TestQueryNodes(t *testing.T) {
	// precondition: a node is registered
	// we can't use the current year, because that would raise the eai rate over 2
	ensureTargetAddressSyncd(t)
	app := initAppRegisterNode(t)
	rn := NewRegisterNode(targetAddress, []byte{0xa0, 0x00, 0x88}, targetPublic, 1, transferPrivate)
	resp := deliverTx(t, app, rn)
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))

	// postcondition: the nodes query returns only that node
	qresp := app.Query(abci.RequestQuery{
		Path: query.NodesEndpoint,
	})
	require.Equal(t, code.OK, code.ReturnCode(qresp.Code))

	var nodes query.NodesResponse
	leftover, err := nodes.UnmarshalMsg(qresp.Value)
	require.NoError(t, err)
	require.Empty(t, leftover)
	require.NotEmpty(t, nodes)

	require.Contains(t, nodes, targetAddress.String())
	require.NotZero(t, nodes[targetAddress.String()])
}

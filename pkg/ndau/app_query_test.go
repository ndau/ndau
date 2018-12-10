package ndau

import (
	"fmt"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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
	app, _, ts := initAppSettlement(t)
	t.Log("timestamp after which escrows expire", ts)
	t.Log("app blocktime", app.blockTime)
	t.Log("comparison", app.blockTime.Compare(ts))
	require.True(t, app.blockTime.Compare(ts) >= 0)

	resp := app.Query(abci.RequestQuery{
		Path: query.AccountEndpoint,
		Data: []byte(settled),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.NotEqual(t, math.Ndau(0), accountData.Balance)
	require.Equal(t, 0, len(accountData.Settlements))
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

	var fee math.Ndau
	n, err := fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee)
	require.NoError(t, err)
	require.Equal(t, 1, n)
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

	var fee math.Ndau
	n, err := fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

func TestCanQuerySidechainTxExists(t *testing.T) {
	sidechainID := byte(0)

	// set up initial state with this sidechain tx paid for
	app, private := initAppTx(t)
	srcA, err := address.Validate(source)
	require.NoError(t, err)

	stx := SidechainTx{
		Source:                 srcA,
		SidechainID:            sidechainID,
		SidechainSignableBytes: []byte{0, 1, 2, 3, 4},
		SidechainSignatures:    nil,
		Sequence:               1,
	}
	stx.Signatures = append(stx.Signatures, metatx.Sign(&stx, private))
	dresp := deliverTx(t, app, &stx)
	require.Equal(t, code.OK, code.ReturnCode(dresp.Code))

	// now set up and run two test cases
	type tcase struct {
		name      string
		hash      string
		wantexist bool
	}
	cases := []tcase{
		{"should exist", stx.SidechainTxHash(), true},
		{"should not exist", "not a real tx hash", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stxq := query.SidechainTxExistsQuery{
				SidechainID: sidechainID,
				Source:      srcA,
				TxHash:      tc.hash,
			}
			qbytes, err := stxq.MarshalMsg(nil)
			require.NoError(t, err)

			resp := app.Query(abci.RequestQuery{
				Path: query.SidechainTxExistsEndpoint,
				Data: qbytes,
			})

			t.Log(code.ReturnCode(resp.Code))
			require.Equal(t, code.OK, code.ReturnCode(resp.Code))
			require.NotEmpty(t, resp.Info)

			var exists bool
			n, err := fmt.Sscanf(resp.Info, query.SidechainTxExistsInfoFmt, &exists)
			require.NoError(t, err)
			require.Equal(t, 1, n)

			require.Equal(t, tc.wantexist, exists)
		})
	}
}

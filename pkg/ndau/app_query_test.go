package ndau

import (
	"fmt"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
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
	var sib math.Ndau
	n, err := fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
	require.Equal(t, 2, n)
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
	var sib math.Ndau
	n, err := fmt.Sscanf(resp.Info, query.PrevalidateInfoFmt, &fee, &sib)
	require.NoError(t, err)
	require.Equal(t, 2, n)
}

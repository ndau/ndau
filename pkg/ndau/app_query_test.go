package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/version"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
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
	require.Equal(t, "acct exists: true", resp.Log)
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
	require.Equal(t, "acct exists: false", resp.Log)
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
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "total ndau at height 1 is 1000000000000, in 1 accounts", resp.Log)
	summary := new(query.Summary)
	_, err := summary.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.BlockHeight)
	require.Equal(t, 1, summary.NumAccounts)
	require.Equal(t, math.Ndau(10000*constants.QuantaPerUnit), summary.TotalNdau)
}

func TestCanQuerySummary2(t *testing.T) {
	app, _ := initAppTx(t)
	// create a new account for different results
	public, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindUser, public.Bytes())
	require.NoError(t, err)
	modify(t, addr.String(), app, func(ad *backing.AccountData) {
		ad.Balance = 1 * constants.QuantaPerUnit
	})
	app.SetHeight(2)

	resp := app.Query(abci.RequestQuery{
		Path: query.SummaryEndpoint,
		Data: nil,
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "total ndau at height 2 is 1000100000000, in 2 accounts", resp.Log)
	summary := new(query.Summary)
	_, err = summary.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, uint64(2), summary.BlockHeight)
	require.Equal(t, 2, summary.NumAccounts)
	require.Equal(t, math.Ndau(10001*constants.QuantaPerUnit), summary.TotalNdau)
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

package ndau

import (
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta.app/code"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestCanQueryAccountStatusSource(t *testing.T) {
	app, _ := initAppTx(t)

	// test the source, which should exist
	resp := app.Query(abci.RequestQuery{
		Path: AccountEndpoint,
		Data: []byte(source),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "exists", resp.Log)
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.Equal(t, math.Ndau(10000*constants.QuantaPerUnit), accountData.Balance)
}

func TestCanQueryAccountStatusDest(t *testing.T) {
	app, _ := initAppTx(t)

	// test the source, which should not exist
	resp := app.Query(abci.RequestQuery{
		Path: AccountEndpoint,
		Data: []byte(dest),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	require.Equal(t, "does not exist", resp.Log)
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
		Path: AccountEndpoint,
		Data: []byte(settled),
	})
	require.Equal(t, code.OK, code.ReturnCode(resp.Code))
	accountData := new(backing.AccountData)
	_, err := accountData.UnmarshalMsg(resp.Value)
	require.NoError(t, err)
	require.NotEqual(t, math.Ndau(0), accountData.Balance)
	require.Equal(t, 0, len(accountData.Settlements))
}

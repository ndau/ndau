package ndau

import (
	"testing"
	"time"

	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppDoesntPanicOnSystemVarTimeout(t *testing.T) {
	app, _ := initApp(t)
	sc := cache.MakeMockCache(t, 50, 10)
	app.systemCache = &sc

	// given the always-timing-out system cache, this used to panic
	// it shouldn't, anymore
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})

	// test the source, which should exist
	resp := app.Query(abci.RequestQuery{
		Path: query.AccountEndpoint,
		Data: []byte(source),
	})
	require.Equal(t, code.InvalidNodeState, code.ReturnCode(resp.Code))
}

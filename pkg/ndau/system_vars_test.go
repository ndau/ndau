package ndau

import (
	"io/ioutil"
	"testing"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
)

func initApp(t *testing.T) (app *App, assc config.MockAssociated) {
	configP, assc, err := config.MakeTmpMock("")
	require.NoError(t, err)

	app, err = NewApp("", *configP)
	require.NoError(t, err)

	// disable logging within the tests by sending output to devnull
	logger := log.StandardLogger()
	logger.Out = ioutil.Discard
	app.SetLogger(logger)

	return
}

// app.System depends on app.Height() returning a reasonable value.
// Also, to test all system variable features, we need to be able to
// control what that value is.
//
// Unfortunately, by default, app.Height just crashes before the app
// is fully initialized, which happens at the InitChain transaction.
//
// We need to send InitChain regardless, to set the initial system
// variable cache,
// but that doesn't allow us to control the returned height, and we
// definitely don't want to wait around for the chain to run for some
// number of blocks.
//
// We've solved this by making what should be a private method, public.
// All we have to do now is call it.
func initAppAtHeight(t *testing.T, atHeight uint64) (app *App) {
	app, _ = initApp(t)
	// adjust only if required
	if atHeight != 0 {
		app.SetHeight(atHeight)
	}
	app.InitChain(types.RequestInitChain{})
	return
}

func testSystem(t *testing.T, app *App, name, expect string) {
	// Note: all these keys/values are presets and defined in
	// config/make_mock.go
	var value wkt.String
	err := app.System(name, &value)
	require.NoError(t, err)
	require.Equal(t, expect, string(value))
}
func TestAppCanGetCurrentValueOfDeferredUpdate(t *testing.T) {
	app := initAppAtHeight(t, 0)
	// this fixture will switch from "bpc val one" to "system value one"
	// at height 1000. Given that we just created this app and haven't
	// run it, we can be confident that it is still at the first value
	testSystem(t, app, "one", "bpc val one")
}

func TestAppCanGetFutureValueOnceHeightIsAppropriate(t *testing.T) {
	app := initAppAtHeight(t, 1000)
	testSystem(t, app, "one", "system value one")
}

func TestAppCanGetSimpleValue(t *testing.T) {
	app := initAppAtHeight(t, 0)
	testSystem(t, app, "two", "system value two")
}

func TestAppCanGetAliasedValue(t *testing.T) {
	app := initAppAtHeight(t, 0)
	testSystem(t, app, "foo", "baz")
}

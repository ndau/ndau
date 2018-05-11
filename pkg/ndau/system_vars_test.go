package ndau

import (
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/oneiro-ndev/ndau-chain/pkg/ndau/config"
	"github.com/stretchr/testify/require"
)

func initApp(t *testing.T) (app *App) {
	configP, err := config.MakeTmpMock("")
	require.NoError(t, err)

	app, err = NewApp("", *configP)
	require.NoError(t, err)

	return
}

// app.System depends on app.Height() returning a reasonable value.
// Also, to test all system variable features, we need to be able to
// control what that value is.
//
// Unfortunately, by default, app.Height just crashes before the app
// is fully initialized, which happens at the InitChain transaction.
//
// We _could_ just send an InitChain transaction to get things going,
// but that doesn't allow us to control the returned height, and we
// definitely don't want to wait around for the chain to run for some
// number of blocks.
//
// Monkey-patching is the answer! However, as this is definitely not
// a normal go feature, it comes with a few more caveats than it would
// in, say, Python. There's a whole list in the library readme:
// https://github.com/bouk/monkey#notes
//
// In short, if these tests start crashing, try running them with -gcflags=-l.
// If that doesn't work, you may need to just skip these tests.
func initAppAtHeight(t *testing.T, atHeight uint64) (app *App) {
	app = initApp(t)
	monkey.PatchInstanceMethod(reflect.TypeOf(app), "Height", func(_ *App) uint64 {
		return atHeight
	})
	return
}

func cleanupApp(app *App) {
	monkey.UnpatchInstanceMethod(reflect.TypeOf(app), "Height")
}

func testSystem(t *testing.T, app *App, name, expect string) {
	// Note: all these keys/values are presets and defined in
	// config/make_mock.go
	value, err := app.System(name)
	require.NoError(t, err)
	require.Equal(t, expect, string(value))
}
func TestAppCanGetCurrentValueOfDeferredUpdate(t *testing.T) {
	app := initAppAtHeight(t, 0)
	defer cleanupApp(app)
	// this fixture will switch from "bpc val one" to "system value one"
	// at height 1000. Given that we just created this app and haven't
	// run it, we can be confident that it is still at the first value
	testSystem(t, app, "one", "bpc val one")
}

func TestAppCanGetFutureValueOnceHeightIsAppropriate(t *testing.T) {
	app := initAppAtHeight(t, 1000)
	defer cleanupApp(app)
	testSystem(t, app, "one", "system value one")
}

func TestAppCanGetSimpleValue(t *testing.T) {
	app := initAppAtHeight(t, 0)
	defer cleanupApp(app)
	testSystem(t, app, "two", "system value two")
}

func TestAppCanGetAliasedValue(t *testing.T) {
	app := initAppAtHeight(t, 0)
	defer cleanupApp(app)
	testSystem(t, app, "foo", "baz")
}

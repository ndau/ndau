package ndau

import (
	"testing"

	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/stretchr/testify/require"
)

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

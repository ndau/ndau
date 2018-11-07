package ndau

import (
	"testing"

	"github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	"github.com/oneiro-ndev/msgp-well-known-types/wkt"
	"github.com/oneiro-ndev/ndau/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	sysvalone signature.PublicKey
	sysvaltwo signature.PrivateKey
	bpcvalone address.Address
	bpcvalbar eai.RTRow
)

func init() {
	var err error
	sysvalone, sysvaltwo, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
	}

	bpcvalone, err = address.Generate(address.KindNdau, []byte("bpc value one"))
	if err != nil {
		panic(err)
	}

	bpcvalbar = eai.RTRow{
		From: math.Duration(math.Day),
		Rate: eai.RateFromPercent(10),
	}
}

func initAppSystem(t *testing.T, height uint64) *App {
	app, _ := initApp(t)

	// inject some test variables
	require.NotNil(t, app.config.UseMock)
	gfile, err := genesisfile.Load(*app.config.UseMock)
	require.NoError(t, err)

	sets := func(ns []byte, key string, value genesisfile.Valuable) {
		tru := true

		loc := svi.Location{
			Namespace: ns,
			Key:       []byte(key),
		}
		gfile.Set(loc, value)
		gfile.Edit(loc, func(val *genesisfile.Value) error {
			val.System = &tru
			return nil
		})
	}

	sets(ns.System, "one", &sysvalone)
	sets(ns.System, "two", &sysvaltwo)

	sviLoc, err := gfile.FindSVIStub()
	require.NoError(t, err)
	require.NotNil(t, sviLoc)
	bpc := sviLoc.Namespace

	sets(bpc, "one", &bpcvalone)
	sets(bpc, "bar", &bpcvalbar)

	// dump the genesisfile
	err = gfile.Dump(*app.config.UseMock)
	require.NoError(t, err)

	// refresh the systemcache
	sc, err := cache.NewSystemCache(app.config)
	require.NoError(t, err)
	app.systemCache = sc

	// update the system cache
	app.InitChain(abci.RequestInitChain{})

	return app
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
	app := initAppSystem(t, 0)
	// this fixture will switch from "bpc val one" to "system value one"
	// at height 1000. Given that we just created this app and haven't
	// run it, we can be confident that it is still at the first value
	testSystem(t, app, "one", "bpc val one")
}

func TestAppCanGetFutureValueOnceHeightIsAppropriate(t *testing.T) {
	app := initAppSystem(t, 1000)
	testSystem(t, app, "one", "system value one")
}

func TestAppCanGetSimpleValue(t *testing.T) {
	app := initAppSystem(t, 0)
	testSystem(t, app, "two", "system value two")
}

func TestAppCanGetAliasedValue(t *testing.T) {
	app := initAppSystem(t, 0)
	testSystem(t, app, "foo", "baz")
}

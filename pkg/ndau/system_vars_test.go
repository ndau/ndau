package ndau

import (
	"testing"

	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	bpcvalone signature.PublicKey
	nsSystem  = []byte("system")
)

func init() {
	var err error
	bpcvalone, _, err = signature.Generate(signature.Ed25519, nil)
	if err != nil {
		panic(err)
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
		err = gfile.Set(loc, value)
		require.NoError(t, err)
		err = gfile.Edit(loc, func(val *genesisfile.Value) error {
			val.System = &tru
			return nil
		})
		require.NoError(t, err)
	}

	sets(nsSystem, "one", &bpcvalone)

	// dump the genesisfile
	err = gfile.Dump(*app.config.UseMock)
	require.NoError(t, err)

	// refresh the systemcache
	// sc, err := cache.NewSystemCache(app.config)
	// require.NoError(t, err)
	// app.systemCache = sc

	// update the system cache
	app.InitChain(abci.RequestInitChain{})

	return app
}

func TestAppCanGetValue(t *testing.T) {
	app := initAppSystem(t, 0)
	// this fixture will switch from "bpc val one" to "system value one"
	// at height 1000. Given that we just created this app and haven't
	// run it, we can be confident that it is still at the first value
	var value signature.PublicKey
	err := app.System("one", &value)
	require.NoError(t, err)
	require.Equal(t, bpcvalone.String(), value.String())
}

// there used to be more tests here, but they depended on the detailed behavior
// of SVI maps. Becuase we are operating in a genesisfile context, and genesisfiles
// always automatically derive SVI maps on load, we can no longer test those
// features. We've therefore just deleted the tests which can no longer run.

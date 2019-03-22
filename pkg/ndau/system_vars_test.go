package ndau

import (
	"testing"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/stretchr/testify/require"
)

var (
	bpcvalone signature.PublicKey
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

	oneb, err := bpcvalone.MarshalMsg(nil)
	require.NoError(t, err)

	app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.Sysvars["one"] = oneb
		return state, nil
	})

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

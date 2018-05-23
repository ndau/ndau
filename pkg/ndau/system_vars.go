package ndau

import (
	"github.com/oneiro-ndev/chaostool/pkg/tool"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	"github.com/pkg/errors"
	trpc "github.com/tendermint/tendermint/rpc/client"
	"github.com/tinylib/msgp/msgp"
)

type chaosClient struct {
	inner trpc.ABCIClient
}

// Static type assertion that chaosClient implements the SystemStore interface
var _ config.SystemStore = (*chaosClient)(nil)

func newChaosClient(address string) chaosClient {
	return chaosClient{
		inner: trpc.NewHTTP(address, "/websocket"),
	}
}

// Get implements the SystemStore interface
func (cc chaosClient) Get(
	namespace []byte,
	key msgp.Marshaler,
	value msgp.Unmarshaler,
) error {
	return tool.GetStructured(
		cc.inner, namespace, key, value, 0,
	)
}

// System retrieves a named system variable.
//
// System variables are normally stored on the chaos chain, so we need
// to query that chain directly most of the time. Because the Blockchain
// Policy Council may want to rename or reassign permissions for these
// variables, there needs to be an indirection layer. Because we want to
// test our code, there needs to be a second indirect where we choose
// whether or not to divert to a mock.
func (app *App) System(name string, value msgp.Unmarshaler) (err error) {
	var ss config.SystemStore
	if len(app.config.UseMock) > 0 {
		ss, err = config.LoadMock(app.config.UseMock)
		if err != nil {
			return errors.Wrap(err, "System() failed to load mock")
		}
	} else {
		ss = newChaosClient(app.config.ChaosAddress)
	}

	svi, err := config.GetSVI(ss, app.config.SystemVariableIndirect)
	if err != nil {
		return errors.Wrap(err, "System() could not find SVI")
	}
	nsk, err := svi.Get(name, app.Height())
	if err != nil {
		return errors.Wrap(err, "System() could not locate desired name")
	}

	return errors.Wrap(
		config.GetNSK(ss, nsk, value),
		"System() could not find named indirect target",
	)
}

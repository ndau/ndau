package ndau

import (
	"github.com/oneiro-ndev/ndau-chain/pkg/ndau/config"
	"github.com/pkg/errors"
	trpc "github.com/tendermint/tendermint/rpc/client"
	"gitlab.ndau.tech/experiments/chaostool/pkg/tool"
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
func (cc chaosClient) Get(namespace, key []byte) (result []byte, err error) {
	result, _, err = tool.GetNamespacedAt(
		cc.inner, namespace, key, 0,
	)
	return
}

// System retrieves a named system variable.
//
// System variables are normally stored on the chaos chain, so we need
// to query that chain directly most of the time. Because the Blockchain
// Policy Council may want to rename or reassign permissions for these
// variables, there needs to be an indirection layer. Because we want to
// test our code, there needs to be a second indirect where we choose
// whether or not to divert to a mock.
func (app *App) System(name string) (val []byte, err error) {
	var ss config.SystemStore
	if len(app.config.UseMock) > 0 {
		ss, err = config.LoadMock(app.config.UseMock)
		if err != nil {
			return nil, errors.Wrap(err, "System() failed to load mock")
		}
	} else {
		ss = newChaosClient(app.config.ChaosAddress)
	}

	svi, err := config.GetSVI(ss, app.config.SystemVariableIndirect)
	if err != nil {
		return nil, errors.Wrap(err, "System() could not find SVI")
	}
	snsk := svi.GetNamespacedKey(name, app.Height())
	if snsk == nil {
		return nil, errors.New("System() could not locate desired name")
	}
	nsk := snsk.AsNamespacedKey()
	val, err = config.GetNSK(ss, nsk)
	return val, errors.Wrap(err, "System() could not find named indirect target")
}

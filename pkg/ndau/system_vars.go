package ndau

import (
	"github.com/oneiro-ndev/ndau-chain/pkg/ndau/config"
	"github.com/pkg/errors"
)

// System retrieves a named system variable.
//
// System variables are normally stored on the chaos chain, so we need
// to query that chain directly most of the time. Because the Blockchain
// Policy Council may want to rename or reassign permissions for these
// variables, there needs to be an indirection layer. Because we want to
// test our code, there needs to be a second indirect where we choose
// whether or not to divert to a mock.
func (app *App) System(name string) ([]byte, error) {
	if len(app.config.UseMock) > 0 {
		return app.systemFromMock(name)
	}
	return app.systemFromChaos(name)
}

func (app *App) systemFromChaos(name string) ([]byte, error) {
	return nil, nil
}

func (app *App) systemFromMock(name string) ([]byte, error) {
	mock, err := config.LoadMock(app.config.UseMock)
	if err != nil {
		return nil, errors.Wrap(err, "System() failed to load mock")
	}
	svi, err := config.GetSVI(mock, app.config.SystemVariableIndirect)
	if err != nil {
		return nil, errors.Wrap(err, "System() could not find SVI")
	}
	snsk := svi.GetNamespacedKey(name, app.Height())
	if snsk == nil {
		return nil, errors.New("System() could not locate desired name")
	}
	nsk := snsk.AsNamespacedKey()
	val, err := config.GetNSK(mock, nsk)
	return val, errors.Wrap(err, "System() could not find named indirect target")
}

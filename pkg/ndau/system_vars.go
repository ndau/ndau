package ndau

import (
	"github.com/tinylib/msgp/msgp"
)

// System retrieves a named system variable.
//
// System variables are normally stored on the chaos chain, so we need
// to query that chain directly most of the time. Because the Blockchain
// Policy Council may want to rename or reassign permissions for these
// variables, there needs to be an indirection layer. Because we want to
// test our code, there needs to be a second indirect where we choose
// whether or not to divert to a mock.
func (app *App) System(name string, value msgp.Unmarshaler) (err error) {
	return app.systemCache.Get(name, value)
}

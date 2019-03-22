package ndau

import (
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// System retrieves a named system variable.
func (app *App) System(name string, value msgp.Unmarshaler) (err error) {
	return errors.New("unimplemented")
}

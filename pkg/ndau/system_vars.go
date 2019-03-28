package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/tinylib/msgp/msgp"
)

// System retrieves a named system variable.
func (app *App) System(name string, value msgp.Unmarshaler) (err error) {
	state := app.GetState().(*backing.State)
	bytes, exists := state.Sysvars[name]
	if !exists {
		return fmt.Errorf("Sysvar %s does not exist", name)
	}
	var leftovers []byte
	leftovers, err = value.UnmarshalMsg(bytes)
	if err == nil && len(leftovers) > 0 {
		err = fmt.Errorf("Sysvar %s has extra trailing bytes; this is suspicious", name)
	}
	return
}

package ndau

import (
	"github.com/tendermint/abci/types"
)

// Update the app's internal state with the given validator
func (app *App) updateValidator(v types.Validator) {
	logger := app.logger.With("method", "updateValidator")
	logger.Info("entered method", "Power", v.GetPower(), "PubKey", v.GetPubKey())
	app.state.UpdateValidator(app.db, v)

	// we only update the changes array after updating the tree
	app.ValUpdates = append(app.ValUpdates, v)
	logger.Info("exiting OK", "app.ValUpdates", app.ValUpdates)
}

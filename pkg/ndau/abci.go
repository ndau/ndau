package ndau

import abci "github.com/tendermint/tendermint/abci/types"

// BeginBlock overrides the metanode BeginBlock ABCI message handler.
//
// If a quit is pending, the application (and the ndaunode executable) exits.
// Otherwise, just uses the default handler.
func (app *App) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	if app.quitPending {
		quit()
	}
	return app.App.BeginBlock(req)
}

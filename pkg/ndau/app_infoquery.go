// This file contains info/query connection methods for the App

package ndau

import (
	"github.com/tendermint/abci/types"
)

// Info services Info requests
func (app *App) Info(req types.RequestInfo) (resInfo types.ResponseInfo) {
	app.logRequest("Info")
	return types.ResponseInfo{
		LastBlockHeight:  int64(app.Height()),
		LastBlockAppHash: app.Hash(),
	}
}

// Query determines the current value for a given key
func (app *App) Query(request types.RequestQuery) (response types.ResponseQuery) {
	app.logRequest("Info")
	response.Height = int64(app.Height())
	return
}

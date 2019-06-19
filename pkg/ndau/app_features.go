package ndau

import (
	"github.com/oneiro-ndev/metanode/pkg/meta/app"
)

// Features allows us to gate ndau-specific feature logic based on app state.
type Features struct {
	app.Features
}

// IsActive returns whether the given feature is currently active in the app.
func (app *App) IsActive(feature string) bool {
	return true
}

// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/abci/types#Application

package ndau

import (
	"time"

	meta "github.com/oneiro-ndev/metanode/pkg/meta.app"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	"github.com/tendermint/abci/types"

	"github.com/pkg/errors"
)

// App is an ABCI application which implements the Ndau chain
type App struct {
	*meta.App
	// configuration data loaded at initialization
	// for now, this just stores the necessary info to get system variables
	// from the chaos chain (or a mock as necessary), but it permits
	// growth as requirements evolve
	config config.Config

	// cache of system variables, updated every block
	systemCache *cache.SystemCache

	// official chain time of the current block
	blockTime math.Timestamp
}

// NewApp prepares a new Ndau App
func NewApp(dbSpec string, config config.Config) (*App, error) {
	metaapp, err := meta.NewApp(dbSpec, "ndau", new(backing.State), TxIDs)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create metaapp")
	}

	sc, err := cache.NewSystemCache(config)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create system variable cache")
	}

	initialBlockTime, err := math.TimestampFrom(constants.Epoch)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create initial block time")
	}

	app := App{
		metaapp,
		config,
		sc,
		initialBlockTime,
	}
	app.App.SetChild(&app)
	return &app, nil
}

// InitChain performs necessary chain initialization.
//
// Most of this is taken care of for us by meta.App, but we
// still need to initialize the system variable cache ourselves
func (app *App) InitChain(req types.RequestInitChain) (response types.ResponseInitChain) {
	// perform basic chain init
	response = app.App.InitChain(req)

	// update system variable cache
	err := app.systemCache.Update(app.Height())
	if err != nil {
		app.GetLogger().Error(
			"failed update of system variable cache",
			"err", err.Error(),
		)
		// given that the system hasn't properly come up yet, I feel no shame
		// simply aborting here
		panic(err)
	}

	return
}

// BeginBlock is called every time a block starts
//
// Most of this is taken care of for us by meta.App, but we need to
// update the current block time.
func (app *App) BeginBlock(req types.RequestBeginBlock) (response types.ResponseBeginBlock) {
	response = app.App.BeginBlock(req)

	header := req.GetHeader()
	tmTime := header.GetTime()
	goTime := time.Unix(tmTime, 0)
	blockTime, err := math.TimestampFrom(goTime)
	if err != nil {
		app.GetLogger().Error(
			"Failed to create ndau timestamp from block time",
			"goTime", goTime,
		)
		panic(err)
	}
	app.blockTime = blockTime

	return response
}

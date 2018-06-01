// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/abci/types#Application

package ndau

import (
	"fmt"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"

	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/spec"
	"github.com/pkg/errors"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"
)

// App is an ABCI application which implements the Ndau chain
type App struct {
	types.BaseApplication

	// We're using noms, which isn't quite like traditional
	// relational databases. In particular, we can't simply
	// store the database, get a cursor, and use the db's stateful
	// nature to keep track of what table we're modifying.
	//
	// Instead, noms breaks things down a bit differentely:
	// the database object manages communication with the server,
	// and most history; the dataset is the working set with
	// which we make updates and then push commits.
	//
	// We therefore need to store both.
	db datas.Database
	ds datas.Dataset

	// Why store this state at all? Why not just have an app.state() function
	// which constructs it in realtime from app.ds.HeadValue?
	//
	// We want to ensure that at all times, the 'official' state committed
	// into the dataset is only updated on a 'Commit' transaction. This
	// in turn means that we need to persist the state between transactions
	// in memory, which means keeping track of this state object.
	state backing.State

	// List of pending validator updates
	ValUpdates []types.Validator

	logger log.Logger

	// configuration data loaded at initialization
	// for now, this just stores the necessary info to get system variables
	// from the chaos chain (or a mock as necessary), but it permits
	// growth as requirements evolve
	config config.Config

	// cache of system variables, updated every block
	systemCache *cache.SystemCache
}

// NewApp prepares a new Ndau App
func NewApp(dbSpec string, config config.Config) (*App, error) {
	if len(dbSpec) == 0 {
		dbSpec = "mem"
	}

	sp, err := spec.ForDatabase(dbSpec)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create noms db")
	}

	var db datas.Database
	// we can fail to connect to noms for a variety of reasons, catch these here and report error
	// we use Try() because noms panics in various places (probably not the right way to handle this)
	err = d.Try(func() {
		db = sp.GetDatabase()
	})
	if err != nil {
		return nil, errors.Wrap(d.Unwrap(err), fmt.Sprintf("NewApp failed to connect to noms db, is noms running at: %s?", dbSpec))
	}

	// in some ways, a dataset is like a particular table in the db
	ds := db.GetDataset("ndau")

	state, ds, err := backing.LoadState(db, ds)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to load existing state")
	}

	sc, err := cache.NewSystemCache(config)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create system variable cache")
	}

	app := App{
		db:          db,
		ds:          ds,
		state:       state,
		logger:      log.NewNopLogger(),
		config:      config,
		systemCache: sc,
	}
	return &app, nil
}

// SetLogger sets the logger to be used by this app
func (app *App) SetLogger(logger log.Logger) {
	app.logger = logger
}

// LogState emits a log message detailing the current app state
func (app *App) LogState() {
	app.logger.Info(
		"LogState",
		"height", app.Height(),
		"hash", app.HashStr(),
	)
}

// logRequest emits a log message on request receipt
//
// It also returns a decorated logger for request-internal logging.
func (app *App) logRequestOptHt(method string, showHeight bool) log.Logger {
	decoratedLogger := app.logger.With(
		"method", method,
	)
	if showHeight {
		decoratedLogger.Info(
			"received request",
			"height", app.Height(),
		)
	} else {
		decoratedLogger.Info("received request")
	}
	return decoratedLogger
}

func (app *App) logRequest(m string) log.Logger {
	return app.logRequestOptHt(m, true)
}

func (app *App) logRequestBare(m string) log.Logger {
	return app.logRequestOptHt(m, false)
}

// Close closes the database connection opened on App creation
func (app *App) Close() error {
	return errors.Wrap(app.db.Close(), "Failed to Close ndau.App")
}

// commit the current application state
//
// This is different from Commit, which processes a Commit Tx!
// However, they're related: think HARD before using this function
// outside of func Commit.
func (app *App) commit() (err error) {
	ds, err := app.state.Commit(app.db, app.ds)
	if err == nil {
		app.ds = ds
	}
	return err
}

// Height returns the current height of the application
func (app *App) Height() uint64 {
	// noms starts counting heights from 1
	// tendermint hates this, and won't reconnect
	// if we do so, because it counts from 0
	return app.ds.HeadRef().Height() - 1
}

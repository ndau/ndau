// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/abci/types#Application

package ndau

import (
	"fmt"

	"github.com/attic-labs/noms/go/d"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/spec"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"
)

// App is an ABCI application which implements immutable storage
// of arbitrary namespaced K-V data on the blockchain.
//
// See the "chaos chain" section of the Epistemology whitepaper
// for details.
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

	// List of pending validator updates
	ValUpdates []types.Validator

	logger log.Logger
}

// NewApp prepares a new Chaos App
func NewApp(dbSpec string) (*App, error) {
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
	ds := db.GetDataset("chaos")

	return &App{
		db:     db,
		ds:     ds,
		logger: log.NewNopLogger(),
	}, nil
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
func (app *App) logRequest(method string) log.Logger {
	decoratedLogger := app.logger.With(
		"method", method,
	)
	decoratedLogger.Info(
		"received request",
		"height", app.Height(),
	)
	return decoratedLogger
}

// Close closes the database connection opened on App creation
func (app *App) Close() error {
	return errors.Wrap(app.db.Close(), "Failed to Close chaos.App")
}

// return the current state of the application
func (app *App) state() nt.Map {
	return app.ds.HeadValue().(nt.Map)
}

// commit the current application state
//
// This is different from Commit, which processes a Commit Tx!
func (app *App) commit() error {
	var err error
	app.ds, err = app.db.CommitValue(app.ds, app.state())
	return err
}

// Height returns the current height of the application
func (app *App) Height() uint64 {
	// noms starts counting heights from 1
	// tendermint hates this, and won't reconnect
	// if we do so, because it counts from 0
	return app.ds.HeadRef().Height() - 1
}

// Update the app's internal state with the given validator
func (app *App) updateValidator(v types.Validator) (err error) {
	logger := app.logger.With("method", "updateValidator")
	logger.Info("entered method", "Power", v.GetPower(), "PubKey", v.GetPubKey())
	if v.Power == 0 {
		logger.Info("attempting to remove validator")
		// TODO: remove validator from internal state

	} else {
		logger.Info("attempting to update validator")
		// TODO: upsert validator into internal state
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)
	logger.Info("exiting OK", "app.ValUpdates", app.ValUpdates)
	return nil
}

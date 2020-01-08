// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

// This file contains the basic definition for the ndau ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package ndau

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jackc/pgx/v4"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	generator "github.com/oneiro-ndev/system_vars/pkg/genesis.generator"
	"github.com/oneiro-ndev/system_vars/pkg/genesisfile"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// App is an ABCI application which implements the Ndau chain
type App struct {
	*meta.App
	// configuration data loaded at initialization
	// for now, this just stores the necessary info to get system variables
	// from the chaos chain (or a mock as necessary), but it permits
	// growth as requirements evolve
	config config.Config

	// quitPending is set to true when a valid ChangeSchema tx is received.
	// It instructs the application to exit with a particular exit code
	// before beginning the next block.
	quitPending bool

	// goodnessFunc enables mocking out the goodness function as required for testing
	// in normal operations, it should always remain the default
	goodnessFunc func(string) (int64, error)
}

// NewApp prepares a new Ndau App
func NewApp(dbSpec string, config config.Config) (*App, error) {
	return NewAppWithLogger(dbSpec, config, nil)
}

// NewAppSilent prepares a new Ndau App which doesn't log
func NewAppSilent(dbSpec string, config config.Config) (*App, error) {
	logger := log.New()
	logger.Out = ioutil.Discard

	return NewAppWithLogger(dbSpec, config, logger)
}

// NewAppWithLogger prepares a new Ndau App with the specified logger
func NewAppWithLogger(dbSpec string, config config.Config, logger log.FieldLogger) (*App, error) {
	metaapp, err := meta.NewAppWithLogger(dbSpec, "ndau", new(backing.State), TxIDs, logger)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create metaapp")
	}

	app := App{
		App:         metaapp,
		config:      config,
		quitPending: false,
	}
	app.goodnessFunc = app.goodnessOf
	app.App.SetChild(&app)

	app.DecoratedLogger().Info("app initialization complete")
	return &app, nil
}

// InitializeDB initializes the indexing database according to the app's configuration
//
// This is separate from app creation because there are cases when we need
// to run a temporary app and don't need or want to bring up the indexing DB.
func (app *App) InitializeDB() error {

	if app.config.PostgresConnection == nil || *app.config.PostgresConnection == "" {
		return errors.New("cannot initialize DB; not configured")
	}

	confstr := *app.config.PostgresConnection

	pgcfg, err := pgx.ParseConfig(confstr)
	if err != nil {
		return errors.Wrap(err, "parsing postgres connection string ("+confstr+")")
	}

	// we always want to connect to the "ndau" db
	pgcfg.Database = "ndau"

	if app.config.PostgresPasswordPath != nil && *app.config.PostgresPasswordPath != "" {
		pwdata, err := ioutil.ReadFile(*app.config.PostgresPasswordPath)
		if err != nil {
			return errors.Wrap(err, "reading postgres password file")
		}
		pgcfg.Password = strings.TrimSpace(string(pwdata))
	}
	idxr, err := search.NewClient(pgcfg, app)
	if err != nil {
		return errors.Wrap(err, "initializing indexer")
	}
	app.SetIndexer(idxr)

	// Log initial indexing in case it takes a long time
	logger := app.DecoratedLogger()
	defer func() {
		if err == nil {
			logger.Info("performed initial blockchain indexing")
		} else {
			logger.WithError(err).Error("problem performing initial blockchain indexing")
		}
	}()

	start := time.Now()

	// Perform initial indexing.
	err = idxr.IndexBlockchain(app.GetDB(), app.GetDS())

	duration := time.Since(start)
	logger = logger.WithField("index.elapsed.ns", duration.Nanoseconds())

	if err != nil {
		return errors.Wrap(err, "performing initial indexing")
	}

	return nil
}

// An IMAArg is an argument to the InitMockApp function
//
// This is used for mock app customization, but it's an advanced feature.
// In general, most users should never have to care about this.
type IMAArg struct {
	Name  string
	Value interface{}
}

// InitMockApp creates an empty test application, which is mainly useful for testing.
//
// This uses a freshly-generated config and an in-memory noms.
func InitMockApp(args ...IMAArg) (app *App, assc generator.Associated, err error) {
	var gfilepath, asscpath string

	gfilepath, asscpath, err = generator.GenerateIn("")
	if err != nil {
		return
	}

	var configfile *os.File
	configfile, err = ioutil.TempFile("", "config.*.toml")
	if err != nil {
		return
	}
	var conf *config.Config
	conf, err = config.LoadDefault(configfile.Name())
	if err != nil {
		return
	}
	// do we have a postgres URI?
	for _, arg := range args {
		switch arg.Name {
		case "dburi":
			if val, ok := arg.Value.(string); ok {
				conf.PostgresConnection = &val
			} else {
				err = fmt.Errorf("dburi expects string value; found %T", arg.Value)
			}
		case "dbpwf":
			if val, ok := arg.Value.(string); ok {
				conf.PostgresPasswordPath = &val
			} else {
				err = fmt.Errorf("dbpwf expects string value; found %T", arg.Value)
			}
		}
	}

	app, err = NewAppSilent("", *conf)
	if err != nil {
		return
	}

	gfile, err := genesisfile.Load(gfilepath)

	// load the genesis state data
	err = app.UpdateStateImmediately(func(stI metast.State) (metast.State, error) {
		state := stI.(*backing.State)
		state.Sysvars, err = gfile.IntoSysvars()
		return state, err
	})

	// now load the appropriate associated data
	_, err = toml.DecodeFile(asscpath, &assc)
	if err != nil {
		return
	}

	return
}

func (app *App) getDefaultRecourseDuration() math.Duration {
	var defaultRecoursePeriod math.Duration
	err := app.System(sv.DefaultRecourseDurationName, &defaultRecoursePeriod)
	if err != nil {
		// if the sysvar doesn't exist or is inaccessable, use 1 hour;
		// this was the default at genesis.
		defaultRecoursePeriod = 1 * math.Hour
		err = nil
	}
	return defaultRecoursePeriod
}

func (app *App) getAccount(addr address.Address) (backing.AccountData, bool) {
	return app.GetState().(*backing.State).GetAccount(addr, app.BlockTime(), app.getDefaultRecourseDuration())
}

// IsFeatureActive returns whether the given feature is currently active.
//
// Once a feature becomes "active", it never becomes "inactive".  We can handle this when
// we add more features that override previous features by checking the newest features first.
//
// For example, say we have a feature in some transaction validation code that rounds a qty:
//
//   qty := math.Round(tx.Qty)
//
// Then later we decided to round to the nearest tenth instead, we would write:
//
//   qty := tx.Qty
//   if app.IsFeatureActive("RoundToTenths") {
//       qty = math.Round(qty*10)/10
//   } else {
//       qty = math.Round(qty)
//   }
//
// Then even later we decide to round to the nearest hundredth, we would write:
//
//   qty := tx.Qty
//   if app.IsFeatureActive("RoundToHundredths") {
//       qty = math.Round(qty*100)/100
//   } else if app.IsFeatureActive("RoundToTenths") {
//       qty = math.Round(qty*10)/10
//   } else {
//       qty = math.Round(qty)
//   }
//
// That way we remain backward compatible until the new rules become active as the app's
// state (i.e. block height) increases.
//
//   height:        0          120               300
//                  |           |                 |
//   blockchain:    |--x---x----+---y------y------+--z--z-------z---...
//                  |           |                 |
//   feature:    genesis   RoundToTenths   RoundToHundredths
//
// A transaction "x" that occurs prior to block 120 gets the default handling since genesis.
// A transaction "y" with height in [120, 300) gets the rounding-by-tenths handling.
// A transaction "z" on or after block height 300 gets the rounding-by-hundredths handling.
func (app *App) IsFeatureActive(feature string) bool {
	// If features is nil, it means that all features are active all the time.
	if app.config.Features == nil {
		return true
	}

	gateHeight, ok := app.config.Features[feature]

	// Unknown or unconfigured features are always active by default.
	if !ok {
		return true
	}

	return app.Height() >= gateHeight
}

// CalculateTxFeeNapu implements AppIndexable
func (app *App) CalculateTxFeeNapu(tx metatx.Transactable) (uint64, error) {
	fee, err := app.calculateTxFee(tx)
	return uint64(fee), err
}

// CalculateTxSIBNapu implements AppIndexable
func (app *App) CalculateTxSIBNapu(tx metatx.Transactable) (uint64, error) {
	ntx, ok := tx.(NTransactable)
	if !ok {
		return 0, fmt.Errorf("%T does not implement NTransactable", tx)
	}
	sib, err := app.calculateSIB(ntx)
	return uint64(sib), err
}

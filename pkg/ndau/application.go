// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package ndau

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
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
}

// NewApp prepares a new Ndau App
func NewApp(dbSpec string, indexAddr string, indexVersion int, config config.Config) (*App, error) {
	return NewAppWithLogger(dbSpec, indexAddr, indexVersion, config, nil)
}

// NewAppSilent prepares a new Ndau App which doesn't log
func NewAppSilent(dbSpec string, indexAddr string, indexVersion int, config config.Config) (*App, error) {
	logger := log.New()
	logger.Out = ioutil.Discard

	return NewAppWithLogger(dbSpec, indexAddr, indexVersion, config, logger)
}

// NewAppWithLogger prepares a new Ndau App with the specified logger
func NewAppWithLogger(dbSpec string, indexAddr string, indexVersion int, config config.Config, logger log.FieldLogger) (*App, error) {
	metaapp, err := meta.NewAppWithLogger(dbSpec, "ndau", new(backing.State), TxIDs, logger)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create metaapp")
	}

	if indexVersion >= 0 {
		// Set up ndau-specific search client.
		search, err := search.NewClient(indexAddr, indexVersion)
		if err != nil {
			return nil, errors.Wrap(err, "NewApp unable to init search client")
		}

		// Log initial indexing in case it takes a long time, people can see why.
		metaapp.GetLogger().WithFields(log.Fields{
			"search.indexVersion": indexVersion,
		}).Info("ndau waiting for initial indexing to complete")

		// Perform initial indexing.
		updateCount, insertCount, err := search.IndexBlockchain(
			metaapp.GetDB(), metaapp.GetDS())
		if err != nil {
			return nil, errors.Wrap(err, "NewApp unable to perform initial indexing")
		}

		// It might be useful to see what kind of results came from the initial indexing.
		metaapp.GetLogger().WithFields(log.Fields{
			"search.updateCount": updateCount,
			"search.insertCount": insertCount,
		}).Info("ndau initial indexing complete")

		metaapp.SetSearch(search)
	}

	app := App{
		metaapp,
		config,
		false,
	}
	app.App.SetChild(&app)

	// We set this environment variable on our devnet nodes to disable feature height gating.
	// The devnet blockchain starts from scratch, so all features are enabled at genesis.
	if os.Getenv("USE_ZERO_HEIGHT_FEATURES") == "1" {
		app.config.Features = nil
	}

	return &app, nil
}

// InitMockApp creates an empty test application, which is mainly useful for testing.
//
// This uses a freshly-generated chaos config and an in-memory noms.
func InitMockApp() (app *App, assc generator.Associated, err error) {
	return InitMockAppWithIndex("", -1)
}

// InitMockAppWithIndex creates an empty test application with indexing and search capability,
// which is mainly useful for testing.
//
// This uses a freshly-generated chaos config and an in-memory noms.
func InitMockAppWithIndex(indexAddr string, indexVersion int) (
	app *App, assc generator.Associated, err error,
) {
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

	app, err = NewAppSilent("", indexAddr, indexVersion, *conf)
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

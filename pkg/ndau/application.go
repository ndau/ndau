// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package ndau

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	sv "github.com/oneiro-ndev/system_vars/pkg/system_vars"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/abci/types"
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

	sc, err := cache.NewSystemCache(config)
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create system variable cache")
	}

	initialBlockTime, err := math.TimestampFrom(time.Now())
	if err != nil {
		return nil, errors.Wrap(err, "NewApp failed to create initial block time")
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
		sc,
		initialBlockTime,
	}
	app.App.SetChild(&app)
	return &app, nil
}

func (app *App) updateSystemVariableCache() error {
	// update system variable cache
	err := app.systemCache.Update(app.Height(), app.GetLogger())
	if err != nil {
		app.GetLogger().WithError(err).Error(
			"failed update of system variable cache",
		)
	}
	// if err == nil, then the state is valid. Otherwise, this blocks us from
	// returning potentially invalid information to callers.
	app.SetStateValidity(err)
	return err
}

// InitChain performs necessary chain initialization.
//
// Most of this is taken care of for us by meta.App, but we
// still need to initialize the system variable cache ourselves
func (app *App) InitChain(req types.RequestInitChain) (response types.ResponseInitChain) {
	// perform basic chain init
	response = app.App.InitChain(req)

	// now wait, potentially forever, for chaos chain (and sysvars)
	sleep := time.Second / 4
	// exponential backoff
	for err := app.updateSystemVariableCache(); err != nil; {
		app.GetLogger().WithError(err).Errorf("trying again after sleep of %s", sleep)
		time.Sleep(sleep)
		sleep *= 2
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
	blockTime, err := math.TimestampFrom(tmTime)
	if err != nil {
		app.GetLogger().WithError(err).WithField("block time", tmTime).Error(
			"failed to create ndau timestamp from block time",
		)
		panic(err)
	}
	app.blockTime = blockTime
	app.updateSystemVariableCache()

	app.GetLogger().WithFields(log.Fields{
		"height": app.Height(),
		"time":   app.blockTime,
	}).Info("ndaunode per block custom processing complete")

	return response
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
	var bpc []byte
	var gfilepath, asscpath string

	bpc, gfilepath, asscpath, err = generator.GenerateIn("")
	if err != nil {
		return
	}

	// update the config with the genesisfile path and the
	// svi location
	var gfile genesisfile.GFile
	gfile, err = genesisfile.Load(gfilepath)
	if err != nil {
		return
	}
	var svi *svi.Location
	svi, err = gfile.FindSVIStub()
	if err != nil {
		return
	}
	if svi == nil {
		err = errors.New("svi stub must exist in generated genesisfile")
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
	conf.UseMock = &gfilepath
	conf.SystemVariableIndirect = *svi
	err = conf.Dump(configfile.Name())
	if err != nil {
		return
	}

	app, err = NewAppSilent("", indexAddr, indexVersion, *conf)
	if err != nil {
		return
	}

	// now load the appropriate associated data
	var af generator.AssociatedFile
	_, err = toml.DecodeFile(asscpath, &af)
	if err != nil {
		return
	}
	var ok bool
	assc, ok = af[base64.StdEncoding.EncodeToString(bpc)]
	if !ok {
		err = errors.New("associated data for this bpc not found in assc file")
		return
	}

	return
}

func (app *App) getDefaultSettlementDuration() math.Duration {
	var defaultSettlementPeriod math.Duration
	err := app.System(sv.DefaultSettlementDurationName, &defaultSettlementPeriod)
	// app.System errors in two cases:
	// - the system variable doesn't exist, which can mean one of two things:
	//   - the SVI map has been updated in an invalid way
	//   - the system cache wasn't updated this block (in which case all txs are
	//     already rejected, so we should never actually see this)
	// - the variable we passed to receive the sysvar is of the wrong type
	//
	// Given this situation, we want to fail in the most noisy way possible.
	if err != nil {
		app.DecoratedLogger().WithError(err).Error("app.getAccount failed to fetch defaultSettlementPeriod")
		panic(err)
	}
	return defaultSettlementPeriod
}

func (app *App) getAccount(addr address.Address) (backing.AccountData, bool) {
	return app.GetState().(*backing.State).GetAccount(addr, app.blockTime, app.getDefaultSettlementDuration())
}

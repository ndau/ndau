// This file contains the basic definition for an ABCI Application.
//
// Interface: https://godoc.org/github.com/tendermint/tendermint/abci/types#Application

package ndau

import (
	"fmt"
	"io/ioutil"
	"time"

	meta "github.com/oneiro-ndev/metanode/pkg/meta/app"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/cache"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/bitset256"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
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

func (app *App) updateSystemVariableCache() {
	// update system variable cache
	err := app.systemCache.Update(app.Height(), app.GetLogger())
	if err != nil {
		app.GetLogger().WithError(err).Error(
			"failed update of system variable cache",
		)
		// given that the system hasn't properly come up yet, I feel no shame
		// simply aborting here
		panic(err)
	}
}

// InitChain performs necessary chain initialization.
//
// Most of this is taken care of for us by meta.App, but we
// still need to initialize the system variable cache ourselves
func (app *App) InitChain(req types.RequestInitChain) (response types.ResponseInitChain) {
	// perform basic chain init
	response = app.App.InitChain(req)

	app.updateSystemVariableCache()

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
	app.updateSystemVariableCache()

	app.GetLogger().WithFields(log.Fields{
		"height": app.Height(),
		"time":   app.blockTime,
	}).Info("ndaunode per block custom processing complete")

	return response
}

// getTxAccount gets and validates an account for a transactable
//
// It returns a nil error if all of:
//  - 1 of N signature validation passes
//  - sequence number is high enough
//  - validation script passes if present
//  - account contains enough ndau to pay the transaction fee
func (app *App) getTxAccount(tx metatx.Transactable, address address.Address, sequence uint64, signatures []signature.Signature) (backing.AccountData, bool, *bitset256.Bitset256, error) {
	validateScript := func(acct backing.AccountData, sigset *bitset256.Bitset256) error {
		if len(acct.ValidationScript) > 0 {
			vm, err := BuildVMForTxValidation(acct.ValidationScript, acct, tx, sigset, app.blockTime)
			if err != nil {
				return errors.Wrap(err, "couldn't build vm for validation script")
			}
			err = vm.Run(false)
			if err != nil {
				return errors.Wrap(err, "validation script")
			}
			vmReturn, err := vm.Stack().PopAsInt64()
			if err != nil {
				return errors.Wrap(err, "validation script exited without numeric stack top")
			}
			if vmReturn != 0 {
				return errors.New("validation script exited with non-0 exit code")
			}
		}
		return nil
	}

	acct, exists, sigset, err := app.GetState().(*backing.State).GetValidAccount(address, app.blockTime, sequence, tx.SignableBytes(), signatures)
	if err != nil {
		return acct, exists, sigset, err
	}

	err = validateScript(acct, sigset)
	if err != nil {
		return acct, exists, sigset, err
	}

	fee, err := app.calculateTxFee(tx)
	if err != nil {
		return acct, exists, sigset, err
	}

	if acct.Balance.Compare(fee) < 0 {
		err = fmt.Errorf("insufficient balance to pay tx fee (%s ndau)", fee)
		return acct, exists, sigset, err
	}

	return acct, exists, sigset, err
}

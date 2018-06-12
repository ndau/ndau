// This file contains consensus connection methods for the App

package ndau

import (
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/code"
	"github.com/tendermint/abci/types"
)

// InitChain performs necessary chain initialization.
//
// This includes saving the initial validator set in the local state.
func (app *App) InitChain(req types.RequestInitChain) (response types.ResponseInitChain) {
	logger := app.logRequestBare("InitChain")

	// now add the initial validators set
	for _, v := range req.Validators {
		app.state.UpdateValidator(app.db, v)
	}

	// commiting here ensures two things:
	// 1. we actually have a head value
	// 2. the initial validators are present from height 1 (or 0, tendermint style)
	err := app.commit()
	if err != nil {
		logger.Error(err.Error())
		// fail fast if we can't actually initialize the chain
		panic(err.Error())
	}

	app.ValUpdates = make([]types.Validator, 0)
	// update system variable cache
	err = app.systemCache.Update(app.Height())
	if err != nil {
		logger.Error(
			"failed update of system variable cache",
			"err", err.Error(),
		)
		// given that the system hasn't properly come up yet, I feel no shame
		// simply aborting here
		panic(err)
	}

	return
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	logger := app.logRequest("BeginBlock")
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	// update system variable cache
	// this is safe to do asynchronously: the first thing an Update does
	// is acquire a lock; it doesn't release the lock until updates are
	// complete. This means that even if we go straight from here to another
	// transaction which requires the values, they'll block until that
	// lock is released.
	go func() {
		err := app.systemCache.Update(app.Height())
		if err != nil {
			// should we panic here? not sure.
			// most of the time we'd expect errors here to be
			// timeouts, and most of the time that won't matter,
			// so I'm inclined to just log it and continue.
			// Might be worth considering for the future, though.
			logger.Error(
				"failed update of system variable cache",
				"err", err.Error(),
			)
		}
	}()
	return types.ResponseBeginBlock{}
}

// DeliverTx services DeliverTx requests
func (app *App) DeliverTx(bytes []byte) (response types.ResponseDeliverTx) {
	app.logRequest("DeliverTx")
	tx, rc, err := app.validateTransactable(bytes)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
		return
	}
	err = tx.Apply(app)
	if err != nil {
		response.Code = uint32(code.ErrorApplyingTransaction)
		response.Log = err.Error()
	}
	return
}

// EndBlock updates the validator set
func (app *App) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	app.logRequest("EndBlock")
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

// Commit saves a new version
//
// Panics if InitChain has not been called.
func (app *App) Commit() types.ResponseCommit {
	logger := app.logRequest("Commit")

	err := app.commit()
	if err != nil {
		logger.Error("Failed to commit block")
		// A panic is appropriate here because the one thing we do _not_ want
		// in the event that a block cannot be committed is for the app to
		// just keep ticking along as if things were ok. Crashing the
		// app should kill the whole node service, which in turn should
		// give human operators a chance to figure out what went wrong.
		//
		// There is no noms documentation stating what kind of errors can
		// be expected from this, but we'd expect them to be mostly I/O
		// issues. In that case, restarting the service, potentially
		// automatically, and recovering state from the rest of the chain
		// is the best way forward.
		panic(err)
	}

	logger.Info("Commit block")
	return types.ResponseCommit{Data: app.Hash()}
}

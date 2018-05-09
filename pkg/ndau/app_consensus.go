// This file contains consensus connection methods for the App

package ndau

import (
	"github.com/oneiro-ndev/ndau-chain/pkg/ndau/code"
	"github.com/tendermint/abci/types"
)

// InitChain saves the validators in the merkle tree
func (app *App) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	app.logRequest("InitChain")
	for _, v := range req.Validators {
		app.updateValidator(v)
	}
	return types.ResponseInitChain{}
}

// BeginBlock tracks the block hash and header information
func (app *App) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	app.logRequest("BeginBlock")
	// reset valset changes
	app.ValUpdates = make([]types.Validator, 0)
	return types.ResponseBeginBlock{}
}

// DeliverTx services DeliverTx requests
// tx a tx.Transaction (defined in tx.Transaction.proto)
func (app *App) DeliverTx(bytes []byte) (response types.ResponseDeliverTx) {
	app.logRequest("DeliverTx")
	tx := new(Transaction)
	err := tx.Unmarshal(bytes)
	if err != nil {
		response.Code = uint32(code.InvalidTransaction)
		response.Log = err.Error()
		return
	}

	nt := ToTransactable(tx)
	if nt == nil {
		response.Code = uint32(code.UnknownTransaction)
		return
	}

	err = nt.Apply(app)
	if err != nil {
		response.Code = uint32(code.ErrorApplyingTransaction)
		response.Log = err.Error()
		return
	}

	response.Code = uint32(code.OK)
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

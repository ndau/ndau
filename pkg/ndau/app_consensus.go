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
		e := app.updateValidator(v)
		if e != nil {
			app.logger.Error("Error updating validators", "error", e.Error())
		}
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
		// Should we do something smarter here?
		// Is there a real way to recover from a failure to save
		// a version?
		// What could cause an error here anyway? TODO:
		// look that up in the noms docs
		panic(err)
	}

	logger.Info("Commit block")
	return types.ResponseCommit{Data: app.Hash()}
}

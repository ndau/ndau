// This file contains mempool connection methods for the App

package ndau

import (
	"github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/code"
	"github.com/tendermint/abci/types"
)

func (app *App) validateTransactable(bytes []byte) (metatx.Transactable, uint32, error) {
	tx, err := metatx.TransactableFromBytes(bytes, TxIDs)
	rc := uint32(code.OK)
	if err != nil {
		app.logger.Info("tx encoding error", "reason", err.Error())
		return nil, uint32(code.EncodingError), err
	}
	err = tx.IsValid(app)
	if err != nil {
		app.logger.Info("invalid tx", "reason", err.Error())
		rc = uint32(code.InvalidTransaction)
		return nil, rc, err
	}
	return tx, rc, nil
}

// CheckTx validates a Transaction
func (app *App) CheckTx(bytes []byte) (response types.ResponseCheckTx) {
	app.logger.Info("Received request", "type", "CheckTx")
	_, rc, err := app.validateTransactable(bytes)
	response.Code = rc
	if err != nil {
		response.Log = err.Error()
	}
	return
}

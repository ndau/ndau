// This file contains mempool connection methods for the App

package ndau

import (
	"github.com/oneiro-ndev/metanode/pkg/meta.transaction"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau/code"
	"github.com/tendermint/abci/types"
)

// CheckTx validates a Transaction
func (app *App) CheckTx(bytes []byte) (response types.ResponseCheckTx) {
	app.logRequest("CheckTx")
	nt, err := metatx.TransactableFromBytes(bytes, TxIDs)
	if err != nil {
		response.Code = uint32(code.InvalidTransaction)
		response.Log = err.Error()
		return
	}

	err = nt.IsValid(app)
	if err != nil {
		response.Code = uint32(code.InvalidTransaction)
		response.Log = err.Error()
		return
	}

	response.Code = uint32(code.OK)
	return
}

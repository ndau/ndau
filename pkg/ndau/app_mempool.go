// This file contains mempool connection methods for the App

package ndau

import (
	"github.com/tendermint/abci/types"
	"gitlab.ndau.tech/experiments/ndau-chain/pkg/ndau/code"
)

// CheckTx validates a Transaction (defined in transaction.proto)
func (app *App) CheckTx(bytes []byte) (response types.ResponseCheckTx) {
	app.logRequest("CheckTx")
	tx := new(Transaction)
	err := tx.Unmarshal(bytes)
	if err != nil {
		response.Code = uint32(code.InvalidTransaction)
		response.Log = err.Error()
		return
	}

	nt := ToNdauTransaction(tx)
	if nt == nil {
		response.Code = uint32(code.UnknownTransaction)
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

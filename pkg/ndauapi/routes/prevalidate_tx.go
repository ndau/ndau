package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// TxJSON is a generic object that contains a completed transaction
// marshaled into []byte and then base-64 encoded into Data as a string.
type TxJSON struct {
	Data string `json:"data"`
}

// PrevalidateResult returns the prevalidation status of a transaction without
// attempting to commit it.
type PrevalidateResult struct {
	FeeNapu int64  `json:"fee_napu"`
	Err     string `json:"err,omitempty"`
	ErrCode int    `json:"err_code,omitempty"`
}

// HandlePrevalidateTx generates a handler that implements the /tx/prevalidate endpoint
func HandlePrevalidateTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// first, get the TxJSON object
		var TxJSON TxJSON

		if r.Body == nil {
			reqres.RespondJSON(w, reqres.NewAPIError("request body required", http.StatusBadRequest))
			return
		}
		err := json.NewDecoder(r.Body).Decode(&TxJSON)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unable to decode", err, http.StatusBadRequest))
			return
		}

		// now decode the transaction
		data, err := base64.StdEncoding.DecodeString(TxJSON.Data)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("tx.Data could not be decoded as base64", err, http.StatusBadRequest))
			return
		}

		tx, err := metatx.Unmarshal(data, ndau.TxIDs)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("tx.Data could not unmarshal into a ndau chain tx", err, http.StatusBadRequest))
			return
		}

		// now we have a tx, prevalidate it
		// first find a node to talk to
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving node", err, http.StatusInternalServerError))
			return
		}

		// and now run the prevalidation query
		fee, _, err := tool.Prevalidate(node, tx)
		result := PrevalidateResult{FeeNapu: int64(fee)}
		code := http.StatusOK // if we ever do this without synchronous commit, change to StatusAccepted
		if err != nil {
			result.Err = err.Error()
			result.ErrCode = -1
			code = http.StatusInternalServerError // probably not the request's fault, actually
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

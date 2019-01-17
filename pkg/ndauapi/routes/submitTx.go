package routes

import (
	"encoding/base64"
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// SubmitResult is returned by the submit endpoint after the tx has been processed
// by a node. If the transaction was valid and posted to the blockchain,
// a TxHash is returned and ResultCode is 0 (and http status will be 200).
// If the included POST body is missing, incorrectly formatted, or cannot be
// submitted to the blockchain as a transaction, this function will
// return 400 as the http status and the SubmitResult return object will not be included.
// If the transaction parses correctly but is determined by the blockchain to be invalid,
// ResultCode will be nonempty and the Msg field will contain a textual error explanation.
// If ResultCode is nonempty, the http status will be 4xx.
// If there was some internal processing error not related to the validity of the
// request or transaction, then http status will be 5xx.
type SubmitResult struct {
	TxHash     string `json:"hash"`
	ResultCode string `json:"result_code,omitempty"`
	Msg        string `json:"msg,omitempty"`
}

// HandleSubmitTx generates a handler that implements the /tx/submit endpoint
func HandleSubmitTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtype := bone.GetValue(r, "txtype")
		mtx, err := TxUnmarshal(txtype, r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("tx.Data did not unmarshal into a tx", err, http.StatusBadRequest))
			return
		}
		tx := mtx.(ndau.NTransactable)

		// now we have a signed tx, submit it
		// first find a node to talk to
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving node", err, http.StatusInternalServerError))
			return
		}

		// and now commit it synchronously
		cr, err := tool.SendCommit(node, tx)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error from commit", err, http.StatusInternalServerError))
			return
		}

		txresult, ok := cr.(*ctypes.ResultBroadcastTxCommit)
		if !ok {
			reqres.RespondJSON(w, reqres.NewFromErr("error casting tx result", err, http.StatusInternalServerError))
			return
		}

		result := SubmitResult{TxHash: base64.StdEncoding.EncodeToString(txresult.Hash)}
		code := http.StatusOK // if we ever do this without synchronous commit, change to StatusAccepted

		if err != nil {
			result.ResultCode = "transaction error"
			result.Msg = err.Error()
			code = http.StatusBadRequest
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

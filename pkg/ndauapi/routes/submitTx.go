package routes

import (
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// SubmitResult is returned by the submit endpoint after the tx has been processed
// by a node. If the transaction was valid and posted to the blockchain,
// a TxHash is returned (and http status will be 200).
// If the included POST body is missing, incorrectly formatted, or cannot be
// submitted to the blockchain as a transaction, this function will
// return 400 as the http status and the SubmitResult return object will not be included.
// If the transaction parses correctly but is determined by the blockchain to be invalid,
// If there was some internal processing error not related to the validity of the
// request or transaction, then http status will be 5xx.
type SubmitResult struct {
	TxHash string `json:"hash"`
	Msg    string `json:"msg,omitempty"`
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

		txhash := metatx.Hash(tx)

		// Check if the tx has already been indexed.
		blockheight, _, err := searchTxHash(node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("txhash search failed", err, http.StatusInternalServerError))
			return
		}

		result := SubmitResult{TxHash: txhash}
		code := http.StatusOK

		// If we've got the tx indexed, it must already be on the blockchain; succeed by default.
		if blockheight > 0 {
			result.Msg = "tx already committed"
			code = http.StatusAccepted
		} else {
			// commit it synchronously; if we ever want to do this asynchronously, we'll need a
			// new endpoint in part because we already use code http.StatusAccepted (202) above.
			cr, err := tool.SendCommit(node, tx)
			if err != nil {
				// chances are high that if this fails, it's the user's fault, so let's
				// blame them, not ourselves
				reqres.RespondJSON(w, reqres.NewFromErr("error from commit", err, http.StatusBadRequest, tool.ResultLog(cr)))
				return
			}
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

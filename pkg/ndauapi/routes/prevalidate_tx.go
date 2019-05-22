package routes

import (
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

// PrevalidateResult returns the prevalidation status of a transaction without
// attempting to commit it.
type PrevalidateResult struct {
	FeeNapu int64  `json:"fee_napu"`
	SibNapu int64  `json:"sib_napu"`
	Err     string `json:"err,omitempty"`
	ErrCode int    `json:"err_code,omitempty"`
	TxHash  string `json:"hash"`
	Msg     string `json:"msg,omitempty"`
}

// HandlePrevalidateTx generates a handler that implements the /tx/prevalidate endpoint
func HandlePrevalidateTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtype := bone.GetValue(r, "txtype")
		tx, err := TxUnmarshal(txtype, r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("tx.Data did not unmarshal into a tx", err, http.StatusBadRequest))
			return
		}

		// now we have a tx, prevalidate it
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
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("txhash search failed: %v", err), http.StatusInternalServerError))
			return
		}

		result := PrevalidateResult{TxHash: txhash}
		code := http.StatusOK

		// If we've got the tx indexed, it must already be on the blockchain; succeed by default.
		if blockheight > 0 {
			result.Msg = "tx already committed"
		} else {
			// run the prevalidation query
			fee, sib, _, err := tool.Prevalidate(node, tx)
			result.FeeNapu = int64(fee)
			result.SibNapu = int64(sib)
			if err != nil {
				result.Err = err.Error()
				result.ErrCode = -1
				code = http.StatusBadRequest
			}
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

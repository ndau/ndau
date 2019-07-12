package routes

import (
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
	FeeNapu          int64  `json:"fee_napu"`
	SibNapu          int64  `json:"sib_napu"`
	ResolveStakeNapu int64  `json:"resolve_stake_napu,omitempty"`
	Err              string `json:"err,omitempty"`
	ErrCode          int    `json:"err_code,omitempty"`
	TxHash           string `json:"hash"`
	Msg              string `json:"msg,omitempty"`
	Code             int    `json:"code"`
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
		vd, err := searchTxHash(node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("txhash search failed", err, http.StatusInternalServerError))
			return
		}
		blockheight := vd.BlockHeight

		result := PrevalidateResult{TxHash: txhash, Code: EndpointResultOK}
		code := http.StatusOK

		// If we've got the tx indexed, it must already be on the blockchain; succeed by default.
		if blockheight > 0 {
			result.Msg = "tx already committed"
			result.Code = EndpointResultTxAlreadyCommitted
			code = http.StatusAccepted
		} else {
			// run the prevalidation query
			fee, sib, resolveStakeCost, _, err := tool.Prevalidate(node, tx)
			result.FeeNapu = int64(fee)
			result.SibNapu = int64(sib)
			result.ResolveStakeNapu = int64(resolveStakeCost)
			if err != nil {
				result.Err = err.Error()
				result.ErrCode = -1
				result.Code = EndpointResultFail
				code = http.StatusBadRequest
			}
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

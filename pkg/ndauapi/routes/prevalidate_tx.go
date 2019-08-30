package routes

import (
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
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
	Code    int    `json:"code"`
}

// HandlePrevalidateTx generates a handler that implements the /tx/prevalidate endpoint
func HandlePrevalidateTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtype := bone.GetValue(r, "txtype")
		tx, err := TxUnmarshal(txtype, r.Body)
		if err != nil {
			cf.Logger.WithError(err).Info("tx.Data did not unmarshal into a tx")
			reqres.RespondJSON(w, reqres.NewFromErr("tx.Data did not unmarshal into a tx", err, http.StatusBadRequest))
			return
		}

		// now we have a tx, prevalidate it
		txhash := metatx.Hash(tx)

		// Check if the tx has already been indexed.
		block, _, _, _, err := searchTxHash(cf.Node, txhash)
		if err != nil {
			cf.Logger.WithError(err).Info("txhash search failed")
			reqres.RespondJSON(w, reqres.NewFromErr("txhash search failed", err, http.StatusInternalServerError))
			return
		}

		result := PrevalidateResult{TxHash: txhash, Code: EndpointResultOK}
		code := http.StatusOK

		// If we've got the tx indexed, it must already be on the blockchain; succeed by default.
		if block != nil {
			result.Msg = "tx already committed"
			result.Code = EndpointResultTxAlreadyCommitted
			code = http.StatusAccepted
		} else {
			// run the prevalidation query
			fee, sib, _, err := tool.Prevalidate(cf.Node, tx, cf.Logger)
			result.FeeNapu = int64(fee)
			result.SibNapu = int64(sib)
			if err != nil {
				cf.Logger.WithError(err).Info("prevalidate returned an error")
				result.Err = err.Error()
				result.ErrCode = -1
				result.Code = EndpointResultFail
				code = http.StatusBadRequest
			}
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

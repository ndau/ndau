package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"

	"github.com/oneiro-ndev/ndau/pkg/tool"

	"github.com/oneiro-ndev/ndaumath/pkg/signature"

	"github.com/oneiro-ndev/ndau/pkg/ndau"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// PreparedTx is a generic object that contains a completed transaction
// marshaled into []byte and then base-64 encoded into TxData as a string.
// Similarly, the SignableBytes is the []byte from the transaction that
// should be signed, again encoded as base-64. When PreparedTx is generated
// on the server side, Signature is not populated, but SignableBytes is.
// When it is received // by the Submit endpoint, it expects an array of 1 or more base-64 encoded
// signatures.
type PreparedTx struct {
	TxData        string
	SignableBytes string
	Signatures    []string
}

// TxResult is returned by the submit endpoint after the tx has been processed
// by a node. If the transaction was valid and posted to the blockchain,
// a TxHash is returned and ResultCode is 0 (and http status will be 200).
// If the included POST body is missing, incorrectly formatted, or cannot be
// submitted to the blockchain as a transaction, this function will
// return 400 as the http status and the TxResult return object will not be included.
// If the transaction parses correctly but is determined by the blockchain to be invalid,
// ResultCode will be nonempty and the ErrorMsg field will contain a textual error explanation.
// If ResultCode is nonempty, the http status will be 4xx.
// If there was some internal processing error not related to the validity of the
// request or transaction, then http status will be 5xx.
type TxResult struct {
	TxHash     string
	ResultCode string
	ErrorMsg   string
}

// HandleSubmitTx generates a handler that implements the /tx/submit endpoint
func HandleSubmitTx(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// first, get the PreparedTx object
		var preparedTx PreparedTx

		if r.Body == nil {
			reqres.RespondJSON(w, reqres.NewAPIError("request body required", http.StatusBadRequest))
			return
		}
		err := json.NewDecoder(r.Body).Decode(&preparedTx)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unable to decode", err, http.StatusBadRequest))
			return
		}

		// now decode the transaction
		data, err := base64.StdEncoding.DecodeString(preparedTx.TxData)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("tx.TxData could not be decoded as base64", http.StatusBadRequest))
			return
		}

		mtx, err := metatx.Unmarshal(data, ndau.TxIDs)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("tx.TxData could not be decoded into a transaction", http.StatusBadRequest))
			return
		}
		tx := mtx.(ndau.NTransactable)

		// see if there are new signatures to add
		if len(preparedTx.Signatures) == 0 {
			reqres.RespondJSON(w, reqres.NewFromErr("at least one signature is required", err, http.StatusBadRequest))
		}

		signable, ok := tx.(ndau.Signable)
		if !ok {
			reqres.RespondJSON(w, reqres.NewAPIError("tx does not implement signable", http.StatusInternalServerError))
			return
		}

		sigs := make([]signature.Signature, len(preparedTx.Signatures))
		for i, s := range preparedTx.Signatures {
			d, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError("could not decode signature as base64", http.StatusInternalServerError))
				return
			}
			_, err = sigs[i].UnmarshalMsg(d)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError("signature could not be decoded", http.StatusInternalServerError))
				return
			}
		}
		signable.AppendSignatures(sigs)

		// now we have a signed tx, submit it
		// first find a node to talk to
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("error retrieving node: %v", err), http.StatusInternalServerError))
			return
		}

		// and now commit it synchronously
		cr, err := tool.SendCommit(node, tx)
		txresult := cr.(*ctypes.ResultBroadcastTxCommit)

		result := TxResult{TxHash: base64.StdEncoding.EncodeToString(txresult.Hash)}
		code := http.StatusOK // if we ever do this without synchronous commit, change to StatusAccepted

		if err != nil {
			result.ResultCode = "transaction error"
			result.ErrorMsg = err.Error()
			code = http.StatusBadRequest
		}

		reqres.RespondJSON(w, reqres.Response{Bd: result, Sts: code})
	}
}

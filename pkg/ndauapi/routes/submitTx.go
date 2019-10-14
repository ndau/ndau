package routes

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
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
	Code   int    `json:"code"`
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
		txhash := metatx.Hash(tx)

		// Check if the tx has already been indexed.
		block, _, _, _, err := searchTxHash(cf.Node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("txhash search failed", err, http.StatusInternalServerError))
			return
		}

		result := SubmitResult{TxHash: txhash, Code: EndpointResultOK}
		code := http.StatusOK

		// If we've got the tx indexed, it must already be on the blockchain; succeed by default.
		if block != nil {
			result.Msg = "tx already committed"
			result.Code = EndpointResultTxAlreadyCommitted
			code = http.StatusAccepted
		} else {
			// commit it synchronously; if we ever want to do this asynchronously, we'll need a
			// new endpoint in part because we already use code http.StatusAccepted (202) above.
			cr, err := tool.SendCommit(cf.Node, tx)
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

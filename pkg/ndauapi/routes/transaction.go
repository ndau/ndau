package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tinylib/msgp/msgp"
)

// TransactionData is the format we use when writing the result of the transaction route.
type TransactionData struct {
	Tx *metatx.Transaction
}

// HandleTransactionFetch gets called by the svc for the /transaction endpoint.
func HandleTransactionFetch(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txhash := bone.GetValue(r, "txhash")
		if txhash == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("txhash parameter required", http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not get node client", http.StatusInternalServerError))
			return
		}

		// Base 64 characters need path-escaping.
		txhash, err = url.PathUnescape(txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not unescape tx hash", http.StatusInternalServerError))
			return
		}

		// Prepare search params.
		params := search.QueryParams{
			Command: search.HeightByTxHashCommand,
			Hash:    txhash,
		}
		paramsBuf := &bytes.Buffer{}
		json.NewEncoder(paramsBuf).Encode(params)
		paramsString := paramsBuf.String()

		searchValue, err := tool.GetSearchResults(node, paramsString)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get search results: %v", err), http.StatusInternalServerError))
			return
		}

		valueData := search.TxValueData{}
		err = valueData.Unmarshal(searchValue)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not parse search results: %v", err), http.StatusInternalServerError))
			return
		}

		blockheight := int64(valueData.BlockHeight)
		if blockheight <= 0 {
			// The search was valid, but there were no results.
			reqres.RespondJSON(w, reqres.OKResponse(nil))
			return
		}

		block, err := node.Block(&blockheight)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not get block: %v", err), http.StatusInternalServerError))
			return
		}

		txoffset := valueData.TxOffset
		if txoffset >= len(block.Block.Data.Txs) {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("tx offset out of range: %d >= %d", txoffset, len(block.Block.Data.Txs)), http.StatusInternalServerError))
			return
		}

		txBytes := block.Block.Data.Txs[txoffset]

		// Use this approach to get the Transaction instead of metatx.Unmarshal() with
		// metatx.AsTransaction() so that we get the same Nonce every time.
		tx := &metatx.Transaction{}
		bytesReader := bytes.NewReader(txBytes)
		msgpReader := msgp.NewReader(bytesReader)
		err = tx.DecodeMsg(msgpReader)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not decode tx: %v", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(tx))
	}
}

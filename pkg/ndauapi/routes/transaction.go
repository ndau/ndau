package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tinylib/msgp/msgp"
)

// TransactionData is the format we use when writing the result of the transaction route.
type TransactionData struct {
	BlockHeight int64
	TxOffset    int
	Tx          *metatx.Transaction
}

func searchTxHash(node *client.HTTP, txhash string) (valueData *search.TxValueData, err error) {
	// Prepare search params.
	params := search.QueryParams{
		Command: search.HeightByTxHashCommand,
		Hash:    txhash,
	}
	paramsBuf := &bytes.Buffer{}
	json.NewEncoder(paramsBuf).Encode(params)
	paramsString := paramsBuf.String()

	var searchValue string
	searchValue, err = tool.GetSearchResults(node, paramsString)
	if err != nil {
		return
	}

	valueData = new(search.TxValueData)
	err = valueData.Unmarshal(searchValue)
	if err != nil {
		return
	}

	return
}

// HandleTransactionFetch gets called by the svc for the /transaction endpoint.
func HandleTransactionFetch(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Transaction hashes are query-escaped by default.
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

		vd, err := searchTxHash(node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("txhash search failed: %v", err), http.StatusInternalServerError))
			return
		}
		blockheight := int64(vd.BlockHeight)
		txoffset := vd.TxOffset

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

		result := TransactionData{
			BlockHeight: blockheight,
			TxOffset:    txoffset,
			Tx:          tx,
		}
		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

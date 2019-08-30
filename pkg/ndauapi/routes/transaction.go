package routes

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tinylib/msgp/msgp"
)

// TransactionData is the format we use when writing the result of the transaction route.
type TransactionData struct {
	BlockHeight int64
	TxOffset    int
	Fee         uint64
	SIB         uint64
	TxType      string
	TxData      metatx.Transactable
}

func searchTxHash(node cfg.TMClient, txhash string) (search.TxValueData, error) {
	// Prepare search params.
	params := search.QueryParams{
		Command: search.HeightByTxHashCommand,
		Hash:    txhash,
	}

	valueData := search.TxValueData{}
	searchValue, err := tool.GetSearchResults(node, params)
	if err != nil {
		return valueData, err
	}

	err = valueData.Unmarshal(searchValue)
	return valueData, err
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

		vd, err := searchTxHash(cf.Node, txhash)
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

		block, err := cf.Node.Block(&blockheight)
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

		txdata, err := tx.AsTransactable(ndau.TxIDs)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not convert tx to transactable: %v", err), http.StatusInternalServerError))
			return
		}

		// Get the transaction type name without the package name on the front.
		txtype := fmt.Sprintf("%T", txdata)
		idx := strings.LastIndex(txtype, ".")
		if idx >= 0 {
			txtype = txtype[idx+1:]
		}

		result := TransactionData{
			BlockHeight: blockheight,
			TxOffset:    txoffset,
			Fee:         vd.Fee,
			SIB:         vd.SIB,
			TxType:      txtype,
			TxData:      txdata,
		}
		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

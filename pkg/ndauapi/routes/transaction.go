package routes

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/tendermint/tendermint/types"
	"github.com/tinylib/msgp/msgp"
)

// TransactionData is the format we use when writing the result of the transaction route.
type TransactionData struct {
	BlockHeight int64
	TxOffset    int
	Fee         uint64
	SIB         uint64
	TxHash      string
	TxType      string
	TxData      metatx.Transactable
}

// TransactionList is the format we use when writing the result of the transaction list route.
// We use an object containing a single array.  This allows us to add more to the response in
// the future without breaking existing clients.
type TransactionList struct {
	Txs        []TransactionData
	NextTxHash string
}

// Search the index for the block containing the transaction with the given hash.
// Returns the block it's in and the tx offset within the block, nil if no search results.
// Also returns the fee and sib values at that transaction.
func searchTxHash(node cfg.TMClient, txhash string) (*types.Block, int, uint64, uint64, error) {
	// Prepare search params.
	params := search.QueryParams{
		Command: search.HeightByTxHashCommand,
		Hash:    txhash,
	}

	valueData := search.TxValueData{}
	searchValue, err := tool.GetSearchResults(node, params)
	if err != nil {
		return nil, -1, 0, 0, err
	}

	err = valueData.Unmarshal(searchValue)
	if err != nil {
		return nil, -1, 0, 0, err
	}
	blockheight := int64(valueData.BlockHeight)
	txoffset := valueData.TxOffset

	if blockheight <= 0 {
		// The search was valid, but there were no results.
		return nil, -1, 0, 0, nil
	}

	block, err := node.Block(&blockheight)
	if err != nil {
		return nil, -1, 0, 0, err
	}

	if txoffset >= int(block.Block.Header.NumTxs) {
		return nil, -1, 0, 0, fmt.Errorf("tx offset out of range: %d >= %d", txoffset, int(block.Block.Header.NumTxs))
	}

	return block.Block, txoffset, valueData.Fee, valueData.SIB, nil
}

// Build a TransactionData out of raw transaction bytes from the blockchain.
func buildTransactionData(txbytes []byte, blockheight int64, txoffset int, txhash string) (*TransactionData, error) {
	// Use this approach to get the Transaction instead of metatx.Unmarshal() with
	// metatx.AsTransaction() so that we get the same Nonce every time.
	tx := &metatx.Transaction{}
	bytesReader := bytes.NewReader(txbytes)
	msgpReader := msgp.NewReader(bytesReader)
	err := tx.DecodeMsg(msgpReader)
	if err != nil {
		return nil, err
	}

	txdata, err := tx.AsTransactable(ndau.TxIDs)
	if err != nil {
		return nil, err
	}

	// Compute the hash if the caller doesn't know it.
	if txhash == "" {
		txhash = metatx.Hash(txdata)
	}

	txtype := metatx.NameOf(txdata)

	return &TransactionData{
		BlockHeight: blockheight,
		TxOffset:    txoffset,
		TxHash:      txhash,
		TxType:      txtype,
		TxData:      txdata,
	}, nil
}

// Start at 'blockheight' (at the tx at 'txoffset' within that block) and fill a list of
// transactions with a max length of 'limit'.
// If txoffset is negative, we start with the latest transaction in the block at blockheight.
func getTransactions(node cfg.TMClient, blockheight int64, txoffset int, limit int) (*TransactionList, error) {
	// This will default to no transactions and no next tx hash.
	result := &TransactionList{}

Loop:
	for h := blockheight; h > 0; h-- {
		block, err := node.Block(&h)
		if err != nil {
			return result, err
		}

		numTxs := int(block.Block.Header.NumTxs)
		// If txoffset is negative, we want to start from the last transaction in this block.
		if txoffset < 0 {
			txoffset = numTxs - 1
		}
		// Work backward through the transaction list since we want reverse chronological order.
		for i := txoffset; i >= 0; i-- {
			transactionData, err := buildTransactionData(block.Block.Data.Txs[txoffset], h, i, "")
			if err != nil {
				return result, err
			}

			// If we've already gotten all the transactions we want, break out of both loops.
			if limit == 0 {
				// This is where the next page of results would start.
				result.NextTxHash = transactionData.TxHash
				break Loop
			}
			limit--

			// Search the index per transaction to get its Fee and SIB values.
			_, _, fee, sib, err := searchTxHash(node, transactionData.TxHash)
			if err != nil {
				return result, err
			}
			transactionData.Fee = fee
			transactionData.SIB = sib

			result.Txs = append(result.Txs, *transactionData)
		}

		// Start with the last transaction in the next block.
		txoffset = -1
	}

	return result, nil
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

		block, txoffset, fee, sib, err := searchTxHash(cf.Node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not find transaction", err, http.StatusInternalServerError))
			return
		}

		// The block will be nil if there were empty search results.
		if block == nil {
			reqres.RespondJSON(w, reqres.OKResponse(nil))
			return
		}

		result, err := buildTransactionData(block.Data.Txs[txoffset], block.Height, txoffset, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not build transaction data", err, http.StatusInternalServerError))
			return
		}

		result.Fee = fee
		result.SIB = sib

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

// HandleTransactionBefore handles requests for transactions on or before a given tx hash.
func HandleTransactionBefore(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Transaction hashes are query-escaped by default.
		txhash := bone.GetValue(r, "txhash")
		if txhash == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("txhash parameter required", http.StatusBadRequest))
			return
		}

		limit, _, err := getPagingParams(r, MaximumRange)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("paging error", err, http.StatusBadRequest))
			return
		}

		qp := getQueryParms(r)
		txtypes := qp["types"]

		if txtypes != "" {
			// TODO: Implement filtering by types using a new index.
			reqres.RespondJSON(w, reqres.OKResponse(&TransactionList{}))
			return
		}

		var blockheight int64
		var txoffset int
		if txhash == "start" {
			// Start with the latest transaction on the blockchain.
			block, err := cf.Node.Block(nil)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("could not get latest block", err, http.StatusInternalServerError))
				return
			}
			blockheight = block.Block.Height
			txoffset = -1 // Negative offset means "start from the latest".
		} else {
			// Find the block and txoffset from which to start gathering a page of transactions.
			var block *types.Block
			block, txoffset, _, _, err = searchTxHash(cf.Node, txhash)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("could not find transaction", err, http.StatusInternalServerError))
				return
			}

			// The block will be nil if there were empty search results.
			if block == nil {
				// Render zero results.  The txoffset is irrelevant in this case.
				blockheight = 0
			} else {
				// We'll render transactions at this height, on or before txoffset.
				blockheight = block.Height
			}
		}

		result, err := getTransactions(cf.Node, blockheight, txoffset, limit)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not get transactions", err, http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

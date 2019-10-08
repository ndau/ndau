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
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/tendermint/tendermint/types"
	"github.com/tinylib/msgp/msgp"
)

// TransactionDataDeprecated is the format we use when writing the result of the transaction route.
type TransactionDataDeprecated struct {
	BlockHeight int64
	TxOffset    int
	Fee         uint64
	SIB         uint64
	Tx          *metatx.Transaction
	TxBytes     []byte
}

// TransactionData is the format we use when writing the result of the transaction route.
type TransactionData struct {
	BlockHeight int64
	TxOffset    int
	Fee         uint64
	SIB         uint64
	TxHash      string
	TxType      string
	TxData      metatx.Transactable
	Timestamp   string
}

// TransactionList is the format we use when writing the result of the transaction list route.
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
func buildTransactionData(timestamp string, txbytes []byte, blockheight int64, txoffset int, txhash string) (*TransactionData, error) {
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
		Timestamp:   timestamp,
		BlockHeight: blockheight,
		TxOffset:    txoffset,
		TxHash:      txhash,
		TxType:      txtype,
		TxData:      txdata,
	}, nil
}

// Search the index for the blocks containing the transaction (or those before it) with the given
// hash and transaction types and page size limit.
// txhash can be one of the following:
//   empty string:   start from the last transaction in the latest block on the blockchain
//   a tx hash:      start from that hash (which might be one of many in a given block)
//   a block height: start from the last transaction in the block on or before that height
// The reason we still need the tool API to take a txhash as opposed to always passing a height
// (which is possible since we have an index that converts txhash to height and txoffset) is
// because we need a height-txoffset pair to avoid having multiple transactions in a block
// appearing on adjacent pages.  This is why we always return a NextTxHash, even if the input
// "hash" is a block height.  It's a way for us to continue where we left off, even if page
// boundaries are mid-block.  We could pass a height-txoffset pair always, but that would require
// us to access the txhash-to-height index here.  Instead, it's done under the hood in tool code.
func searchTxTypes(node cfg.TMClient, txhash string, typeNames []string, limit int) (*TransactionList, error) {
	result := &TransactionList{}

	params := search.QueryParams{
		Command: search.HeightsByTxTypesCommand,
		Hash:    txhash,
		Types:   typeNames,
		Limit:   limit,
	}

	valueData := search.TxListValueData{}
	searchValue, err := tool.GetSearchResults(node, params)
	if err != nil {
		return result, err
	}

	err = valueData.Unmarshal(searchValue)
	if err != nil {
		return result, err
	}

	for i := 0; i < len(valueData.Txs); i++ {
		txdata := valueData.Txs[i]

		blockheight := int64(txdata.BlockHeight)
		block, err := node.Block(&blockheight)
		if err != nil {
			return result, err
		}

		timestamp := block.Block.Header.Time.Format(constants.TimestampFormat)
		txoffset := txdata.TxOffset
		txbytes := block.Block.Data.Txs[txoffset]
		transactionData, err := buildTransactionData(timestamp, txbytes, blockheight, txoffset, "")
		if err != nil {
			return result, err
		}

		transactionData.Fee = txdata.Fee
		transactionData.SIB = txdata.SIB

		result.Txs = append(result.Txs, *transactionData)
	}

	result.NextTxHash = valueData.NextTxHash

	return result, nil
}

// HandleTransactionFetchDeprecated gets called by the svc for the /transaction endpoint.
func HandleTransactionFetchDeprecated(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Transaction hashes are query-escaped by default.
		txhash := bone.GetValue(r, "txhash")
		if txhash == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("txhash parameter required", http.StatusBadRequest))
			return
		}

		block, txoffset, fee, sib, err := searchTxHash(cf.Node, txhash)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("txhash search failed: %v", err), http.StatusInternalServerError))
			return
		}

		if block == nil {
			// The search was valid, but there were no results.
			reqres.RespondJSON(w, reqres.OKResponse(nil))
			return
		}

		if txoffset >= len(block.Data.Txs) {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("tx offset out of range: %d >= %d", txoffset, len(block.Data.Txs)), http.StatusInternalServerError))
			return
		}

		txBytes := block.Data.Txs[txoffset]

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

		result := TransactionDataDeprecated{
			BlockHeight: block.Height,
			TxOffset:    txoffset,
			Fee:         fee,
			SIB:         sib,
			Tx:          tx,
			TxBytes:     txBytes,
		}
		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
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

		timestamp := block.Header.Time.Format(constants.TimestampFormat)
		txbytes := block.Data.Txs[txoffset]
		result, err := buildTransactionData(timestamp, txbytes, block.Height, txoffset, txhash)
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

		// Treat "start" and "" as the same thing: get the latest page of transactions.
		if txhash == "start" {
			txhash = ""
		}

		limit, _, err := getPagingParams(r, MaximumRange)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("paging error", err, http.StatusBadRequest))
			return
		}

		// If no types were given, this will find all transactions.
		typeNames, _ := r.URL.Query()["type"]
		result, err := searchTxTypes(cf.Node, txhash, typeNames, limit)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("could not find transactions", err, http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/constants"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// AccountHistoryItem is used by the account history endpoint to return balance historical data.
type AccountHistoryItem struct {
	Balance   types.Ndau
	Timestamp string
	TxHash    string
	Height    int64
}

// AccountHistoryItems is used by the account history endpoint to return balance historical data.
type AccountHistoryItems struct {
	Items []AccountHistoryItem
	Next  string
}

// HandleAccount returns a HandlerFunc that returns information about a single account
// specified in the URL.
func HandleAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := bone.GetValue(r, "address")
		addrs := []string{addr}
		processAccounts(w, cf.Node, addrs)
	}
}

// HandleAccounts returns a HandlerFunc that returns a collection of account data based
// on a POSTed list of account IDs.
func HandleAccounts(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not read request body", http.StatusBadRequest))
			return
		}
		var addrs []string
		err = json.Unmarshal(bodyBytes, &addrs)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not parse request body as json", http.StatusBadRequest))
			return
		}
		processAccounts(w, cf.Node, addrs)
	}
}

func processAccounts(w http.ResponseWriter, node cfg.TMClient, addresses []string) {
	addies := []address.Address{}
	invalidAddies := []string{}

	for _, one := range addresses {
		a, err := address.Validate(one)
		if err != nil {
			invalidAddies = append(invalidAddies, one)
		}
		addies = append(addies, a)
	}
	if len(invalidAddies) > 0 {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not validate addresses: %v", invalidAddies), http.StatusBadRequest))
		return
	}

	resp := make(map[string]backing.AccountData)
	for _, oneAddy := range addies {
		ad, queryResult, err := tool.GetAccount(node, oneAddy)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address data: %s", err), http.StatusInternalServerError))
			return
		}
		// Check to see if the account was found on the blockchain or if this record was created
		// because it wasn't found. Only create responses for accounts that were found.
		var exists bool
		_, err = fmt.Sscanf(queryResult.Response.Info, query.AccountInfoFmt, &exists)
		if err != nil {
			// if it didn't scan we probably have a compatibility or version issue
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("addr fetch (%s) didn't scan: %s", queryResult.Response.Info, err), http.StatusInternalServerError))
			return
		}
		if exists {
			resp[oneAddy.String()] = *ad
		}
	}
	reqres.RespondJSON(w, reqres.OKResponse(resp))
}

// HandleAccountHistory returns a HandlerFunc that returns balance history about a single account
// specified in the URL.
func HandleAccountHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addressString := bone.GetValue(r, "address")
		if addressString == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("address parameter required", http.StatusBadRequest))
			return
		}

		addr, err := address.Validate(addressString)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not validate address: %s", err), http.StatusBadRequest))
			return
		}

		limit, afters, err := getPagingParams(r, 100)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("paging parms", err, http.StatusBadRequest))
			return
		}

		after := uint64(0)
		if afters != "" {
			after, err = strconv.ParseUint(afters, 10, 64)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("parsing 'after'", err, http.StatusBadRequest))
				return
			}
		}

		// Prepare search params.
		params := search.AccountHistoryParams{
			Address:     addr.String(),
			Limit:       limit,
			AfterHeight: after,
		}

		ahr, _, err := tool.GetAccountHistory(cf.Node, params)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address history: %s", err), http.StatusInternalServerError))
			return
		}

		result := AccountHistoryItems{}

		for _, valueData := range ahr.Txs {
			blockheight := int64(valueData.BlockHeight)
			txoffset := valueData.TxOffset
			balance := valueData.Balance

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

			txab, err := metatx.Unmarshal(txBytes, ndau.TxIDs)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not decode tx: %v", err), http.StatusInternalServerError))
				return
			}

			txhash := metatx.Hash(txab)
			item := AccountHistoryItem{
				Balance:   balance,
				Timestamp: block.Block.Header.Time.Format(constants.TimestampFormat),
				TxHash:    txhash,
				Height:    blockheight,
			}
			result.Items = append(result.Items, item)
		}

		if ahr.More && len(result.Items) > 0 {
			next, err := url.Parse(".")
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("could not parse identity url", err, http.StatusInternalServerError))
				return
			}
			query := r.URL.Query()
			query.Set("after", fmt.Sprint(result.Items[len(result.Items)-1].Height))
			next.RawQuery = query.Encode()
			result.Next = r.URL.ResolveReference(next).String()
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

// HandleAccountList returns a HandlerFunc that returns all the accounts
// in the system
func HandleAccountList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, after, err := getPagingParams(r, 100)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("reading paging info", err, http.StatusBadRequest))
			return
		}

		accts, _, err := tool.GetAccountList(cf.Node, after, limit)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address list: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(accts))
	}
}

// HandleAccountCurrencySeats returns a HandlerFunc that returns all the accounts
// in the system that exceed 1000 ndau; they are sorted in order from oldest
// to newest. It accepts a single parameter for the maximum number of accounts
// to return (default 3000).
func HandleAccountCurrencySeats(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 3000 // the number of currency seats eligible to vote in the 2nd tier election
		qp := getQueryParms(r)
		limitStr := qp["limit"]
		if limitStr != "" {
			pi, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError("limit must be a valid number", http.StatusBadRequest))
				return
			}
			limit = int(pi)
		}

		accts, err := tool.GetCurrencySeats(cf.Node)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching currency seats: %s", err), http.StatusInternalServerError))
			return
		}

		if limit < len(accts) {
			accts = accts[:limit]
		}
		reqres.RespondJSON(w, reqres.OKResponse(accts))
	}
}

package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-zoo/bone"
	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/search"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

// AccountHistoryItem is used by the account history endpoint to return balance historical data.
type AccountHistoryItem struct {
	Balance   types.Ndau
	Timestamp string
	TxHash    string
}

// AccountHistoryItems is used by the account history endpoint to return balance historical data.
type AccountHistoryItems struct {
	Items []AccountHistoryItem
}

// HandleAccount returns a HandlerFunc that returns information about a single account
// specified in the URL.
func HandleAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addr := bone.GetValue(r, "address")
		addrs := []string{addr}
		processAccounts(w, cf.NodeAddress, addrs)
	}
}

// HandleAccounts returns a HandlerFunc that returns a collection of account data based
// on a POSTed list of account IDs.
func HandleAccounts(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		addrs := []string{}
		err := decoder.Decode(&addrs)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not parse request body as json", http.StatusBadRequest))
			return
		}
		processAccounts(w, cf.NodeAddress, addrs)
	}
}

func processAccounts(w http.ResponseWriter, nodeAddr string, addresses []string) {
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

	node, err := ws.Node(nodeAddr)
	if err != nil {
		reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error getting node: %s", err), http.StatusInternalServerError))
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

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error getting node: %s", err), http.StatusInternalServerError))
			return
		}

		pageIndex, pageSize, errMsg, err := getPagingParams(r)
		if errMsg != "" {
			reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
			return
		}

		// Prepare search params.
		params := search.AccountHistoryParams{
			Address:   addr.String(),
			PageIndex: pageIndex,
			PageSize:  pageSize,
		}
		paramsBuf := &bytes.Buffer{}
		json.NewEncoder(paramsBuf).Encode(params)
		paramsString := paramsBuf.String()

		ahr, _, err := tool.GetAccountHistory(node, paramsString)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address history: %s", err), http.StatusInternalServerError))
			return
		}

		result := AccountHistoryItems{}

		for _, valueData := range ahr.Txs {
			blockheight := int64(valueData.BlockHeight)
			txoffset := valueData.TxOffset
			balance := valueData.Balance

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

			txab, err := metatx.Unmarshal(txBytes, ndau.TxIDs)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not decode tx: %v", err), http.StatusInternalServerError))
				return
			}

			txhash := metatx.Hash(txab)
			item := AccountHistoryItem{
				Balance:   balance,
				Timestamp: block.Block.Header.Time.Format(time.RFC3339),
				TxHash:    txhash,
			}
			result.Items = append(result.Items, item)
		}

		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

// HandleAccountList returns a HandlerFunc that returns all the accounts
// in the system
func HandleAccountList(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error getting node: %s", err), http.StatusInternalServerError))
			return
		}

		pageIndex, pageSize, errMsg, err := getPagingParams(r)
		if errMsg != "" {
			reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
			return
		}

		accts, _, err := tool.GetAccountList(node, pageIndex, pageSize)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address list: %s", err), http.StatusInternalServerError))
			return
		}

		reqres.RespondJSON(w, reqres.OKResponse(accts))
	}
}

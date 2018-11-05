package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

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

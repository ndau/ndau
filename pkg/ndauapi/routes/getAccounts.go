package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
)

// AccountDataResponse represents a single account data.
type AccountDataResponse struct {
	backing.AccountData
	Address string `json:"address"`
}

// AccountResponse represents an account response.
type AccountResponse struct {
	AcctData []AccountDataResponse `json:"addressData,omitempty"`
}

// AddressRequest represents a request for a list of addresses.
type AddressRequest struct {
	Addresses []string `json:"addresses"`
}

// GetAccount returns a HandlerFunc that returns an account.
func GetAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		req := AddressRequest{}
		err := decoder.Decode(&req)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("could not parse request body as json", http.StatusBadRequest))
		}

		addies := []address.Address{}
		invalidAddies := []string{}

		for _, one := range req.Addresses {
			a, err := address.Validate(one)
			if err != nil {
				invalidAddies = append(invalidAddies, one)
			}
			addies = append(addies, a)
		}
		if len(invalidAddies) > 0 {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("could not validate addresses: [%v]: %s", invalidAddies, err), http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error getting node: %s", err), http.StatusInternalServerError))
			return
		}

		resp := AccountResponse{}
		for _, oneAddy := range addies {
			ad, _, err := tool.GetAccount(node, oneAddy)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewAPIError(fmt.Sprintf("Error fetching address data: %s", err), http.StatusInternalServerError))
				return
			}
			one := AccountDataResponse{AccountData: *ad}
			one.Address = oneAddy.String()
			resp.AcctData = append(resp.AcctData, one)
		}
		reqres.RespondJSON(w, reqres.OKResponse(resp))
	}
}

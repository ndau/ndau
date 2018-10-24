package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/oneiro-ndev/chaincode/pkg/vm"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

// TxClaimAccountRequest is the format the API expects for the /account/claim endpoint
type TxClaimAccountRequest struct {
	Target           address.Address       `json:"target"`
	OwnershipKey     signature.PublicKey   `json:"ownership"`
	ValidationKeys   []signature.PublicKey `json:"keys"`
	ValidationScript string                `json:"script"`
	Sequence         uint64                `json:"seq"`
}

// HandleClaimAccount is the handler for a claim account request
func HandleClaimAccount(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// first, get the request object
		var acctreq TxClaimAccountRequest

		if r.Body == nil {
			reqres.RespondJSON(w, reqres.NewAPIError("request body required", http.StatusBadRequest))
			return
		}
		err := json.NewDecoder(r.Body).Decode(&acctreq)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("unable to decode", err, http.StatusBadRequest))
			return
		}

		ca := ndau.ClaimAccount{
			Target:         acctreq.Target,
			Ownership:      acctreq.OwnershipKey,
			ValidationKeys: acctreq.ValidationKeys,
			Sequence:       acctreq.Sequence,
		}

		// now decode the script, if there is one
		if acctreq.ValidationScript != "" {
			script, err := base64.StdEncoding.DecodeString(acctreq.ValidationScript)
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("validation script could not be decoded as base64", err, http.StatusBadRequest))
				return
			}

			// and we check it for validity before we store it
			opcodes := vm.ToChaincode(script)
			err = opcodes.IsValid()
			if err != nil {
				reqres.RespondJSON(w, reqres.NewFromErr("ValidationScript is not valid chaincode", err, http.StatusBadRequest))
				return
			}
			ca.ValidationScript = script
		}

		txdata, err := b64Tx(&ca)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("tx could not be marshaled to base64", err, http.StatusInternalServerError))
			return
		}
		preparedTx := PreparedTx{
			TxData:        txdata,
			SignableBytes: b64(ca.SignableBytes()),
		}
		reqres.RespondJSON(w, reqres.Response{Bd: preparedTx, Sts: http.StatusOK})
	}
}

package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	chquery "github.com/oneiro-ndev/chaos/pkg/chaos/query"
	chtool "github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/query"
)

// ChaosItem expresses a single Key/Value on the chaos chain
type ChaosItem struct {
	Key   string
	Value string
}

// ChaosAllResult represents chaos chain data including namespace information
type ChaosAllResult struct {
	Namespace string
	Data      []ChaosItem
}

// SystemHistoryResponse represents the result of querying system history
type SystemHistoryResponse struct {
}

// ChaosHistoryResponse contains the value history of a given key
type ChaosHistoryResponse struct {
	*chquery.KeyHistoryResponse
}

// HandleSystemAll retrieves all the system keys at the current block height.
func HandleSystemAll(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node, err := ws.Node(cf.NodeAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("Could not get a node.", http.StatusInternalServerError))
			return
		}

		resp, err := node.ABCIQuery(query.SysvarsEndpoint, []byte{})
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("Query error: ", err, http.StatusInternalServerError))
			return
		}

		// resp.Value is actually JSON so decode it
		values := make(map[string][]byte)
		err = json.NewDecoder(bytes.NewReader(resp.Response.Value)).Decode(&values)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("Decode error: ", err, http.StatusInternalServerError))
			return
		}
		reqres.RespondJSON(w, reqres.OKResponse(values))
	}
}

// HandleChaosNamespaceAll retrieves all the values in a given namespace at the current block height
func HandleChaosNamespaceAll(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// find the chaos node
		chnode, err := ws.Node(cf.ChaosAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
			return
		}

		// This returns the value already query-unescaped.
		nskey64 := bone.GetValue(r, "namespace")
		nskey, err := base64.StdEncoding.DecodeString(nskey64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr(fmt.Sprintf("error decoding namespace '%s'", nskey64), err, http.StatusBadRequest))
			return
		}

		fmt.Printf("ns key %s, %x\n", nskey64, nskey)
		resp, _, err := chtool.DumpNamespacedAt(chnode, nskey, 0)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos data", err, http.StatusInternalServerError))
			return
		}
		fmt.Printf("%#v\n", resp)
		reqres.RespondJSON(w, reqres.OKResponse(resp))
	}
}

// HandleChaosNamespaceKey retrieves a single namespace value at a given block height
func HandleChaosNamespaceKey(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

// HandleChaosHistory retrieves the history of a single value in the chaos chain.
func HandleChaosHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// This returns the value already query-unescaped.
		namespaceBase64 := bone.GetValue(r, "namespace")
		if namespaceBase64 == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("namespace parameter required", http.StatusBadRequest))
			return
		}

		// This returns the value already query-unescaped.
		keyBase64 := bone.GetValue(r, "key")
		if keyBase64 == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("key parameter required", http.StatusBadRequest))
			return
		}

		namespaceBytes, err := base64.StdEncoding.DecodeString(namespaceBase64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr(fmt.Sprintf("error decoding namespace '%s'", namespaceBase64), err, http.StatusBadRequest))
			return
		}

		keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr(fmt.Sprintf("error decoding key '%s'", keyBase64), err, http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.ChaosAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
			return
		}

		pageIndex, pageSize, errMsg, err := getPagingParams(r)
		if errMsg != "" {
			reqres.RespondJSON(w, reqres.NewFromErr(errMsg, err, http.StatusBadRequest))
			return
		}

		hkr, _, err := chtool.HistoryNamespaced(
			node, namespaceBytes, keyBytes, pageIndex, pageSize)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos data", err, http.StatusInternalServerError))
			return
		}

		result := ChaosHistoryResponse{hkr}
		reqres.RespondJSON(w, reqres.OKResponse(result))
	}
}

// HandleSystemHistory retrieves the history of a single system variable.
func HandleSystemHistory(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

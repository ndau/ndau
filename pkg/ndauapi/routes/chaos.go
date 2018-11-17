package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-zoo/bone"
	cns "github.com/oneiro-ndev/chaos/pkg/chaos/ns"
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

func getSystemValue(cf cfg.Cfg, key string) (string, error) {
	// find the chaos node
	chnode, err := ws.Node(cf.ChaosAddress)
	if err != nil {
		return "", err
	}

	systemns := cns.System
	resp, _, err := chtool.GetNamespacedAt(chnode, systemns, []byte(key), 0)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

// HandleSystemKey retrieves a single system key at the current block height
func HandleSystemKey(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := bone.GetValue(r, "key")
		value, err := getSystemValue(cf, key)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos data", err, http.StatusInternalServerError))
			return
		}

		result := ChaosAllResult{
			Namespace: "system",
			Data: []ChaosItem{ChaosItem{
				Key:   key,
				Value: value,
			}},
		}
		reqres.RespondJSON(w, reqres.OKResponse(result))
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

		nskey64, err := url.PathUnescape(bone.GetValue(r, "namespace"))
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error unescaping namespace", err, http.StatusBadRequest))
			return
		}
		nskey, err := base64.StdEncoding.DecodeString(nskey64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error decoding namespace", err, http.StatusBadRequest))
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
		namespaceBase64Esc := bone.GetValue(r, "namespace")
		if namespaceBase64Esc == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("namespace parameter required", http.StatusBadRequest))
			return
		}

		keyEsc := bone.GetValue(r, "key")
		if keyEsc == "" {
			reqres.RespondJSON(w, reqres.NewAPIError("key parameter required", http.StatusBadRequest))
			return
		}

		namespaceBase64, err := url.PathUnescape(namespaceBase64Esc)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error unescaping namespace", err, http.StatusInternalServerError))
			return
		}

		key, err := url.PathUnescape(keyEsc)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error unescaping key", err, http.StatusInternalServerError))
			return
		}

		namespaceBytes, err := base64.StdEncoding.DecodeString(namespaceBase64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error decoding namespace", err, http.StatusBadRequest))
			return
		}

		node, err := ws.Node(cf.ChaosAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
			return
		}

		hkr, _, err := chtool.HistoryNamespaced(node, namespaceBytes, []byte(key))
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

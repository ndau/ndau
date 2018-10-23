package routes

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	cns "github.com/oneiro-ndev/chaos/pkg/chaos/ns"
	chtool "github.com/oneiro-ndev/chaos/pkg/tool"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
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

// HandleChaosSystemAll retrieves all the system keys at the current block height.
func HandleChaosSystemAll(cf cfg.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// find the chaos node
		chnode, err := ws.Node(cf.ChaosAddress)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos node", err, http.StatusInternalServerError))
			return
		}

		systemns := cns.System
		resp, _, err := chtool.DumpNamespacedAt(chnode, systemns, 0)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error retrieving chaos data", err, http.StatusInternalServerError))
			return
		}

		result := ChaosAllResult{Namespace: "system"}
		for _, r := range (*resp).Data {
			// TODO: Handle Values other than strings?
			result.Data = append(result.Data, ChaosItem{Key: string(r.Key), Value: string(r.Value)})
		}
		reqres.RespondJSON(w, reqres.OKResponse(result))
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

// HandleChaosSystemKey retrieves a single system key at the current block height
func HandleChaosSystemKey(cf cfg.Cfg) http.HandlerFunc {
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

		nskey64 := bone.GetValue(r, "namespace")
		nskey, err := base64.StdEncoding.DecodeString(nskey64)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewFromErr("error decoding key", err, http.StatusBadRequest))
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
	}
}
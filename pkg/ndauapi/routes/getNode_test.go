package routes_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/oneiro-ndev/ndau/pkg/tool"

	"github.com/go-zoo/bone"
)

func TestGetNode(t *testing.T) {

	if os.Getenv("CI") == "true" {
		// early exit for integration type tests
		return
	}

	baseHandler := routes.GetNode

	// set up apparatus
	cf, _, _ := cfg.New()
	handler := baseHandler(cf)
	mux := bone.New()
	mux.Get("/:id", handler)

	// get node first
	node, _ := ws.Node(cf.NodeAddress)
	n, err := tool.Info(node)
	if err != nil {
		t.Error("Couldn't get node.")
		return
	}
	nodeID := n.NodeInfo.ID
	fmt.Println("Node ID:", nodeID)
	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "good request",
			req:    httptest.NewRequest("GET", fmt.Sprintf("/%s", nodeID), nil),
			status: http.StatusOK,
		},
	}
	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
			}
		})
	}

	badHandler := baseHandler(cfg.Cfg{})
	t.Run("empty config", func(t *testing.T) {
		w := httptest.NewRecorder()
		badHandler(w, httptest.NewRequest("GET", "/", nil))
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			body, _ := ioutil.ReadAll(res.Body)
			t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, http.StatusInternalServerError, body)
		}
	})
}

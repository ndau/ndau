package routes_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
)

func TestHandleAccounts(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	baseHandler := routes.HandleAccounts

	// set up tests
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{
			name:   "empty request",
			body:   "[]",
			status: http.StatusOK,
		},
		// Can't really do this test without mocking an address first
		//{
		//	name:   "good address",
		//	body:   "{\"addresses\":[\"\"]}",
		//	status: http.StatusOK,
		//},
		{
			name:   "invalid address",
			body:   "[\"asdf\"]}",
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid json",
			body:   "{\"addresses\"}:[\"asdf\"]}",
			status: http.StatusBadRequest,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
	handler := baseHandler(cf)

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte(tt.body)))
			req.Header.Add("content-type", "application/json")
			handler(w, req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
			}
		})
	}
}

func TestAccountHistory(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
	mux := svc.New(cf).Mux()

	// Add to the blockchain and index.
	addr := createNdauBlock(t)

	// Invalid in that it's not in the index.
	invalidAddr := "invalidaddr_"

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "no address",
			req:      httptest.NewRequest("GET", "/account/history/", nil),
			status:   http.StatusBadRequest,
			wantbody: "address parameter required",
		}, {
			name:     "invalid address",
			req:      httptest.NewRequest("GET", "/account/history/" + invalidAddr, nil),
			status:   http.StatusBadRequest,
			wantbody: "could not validate address",
		}, {
			name:     "valid hash",
			req:      httptest.NewRequest("GET", "/account/history/" + addr, nil),
			status:   http.StatusOK,
			// The addr isn't part of the response, just make sure the response is non-empty.
			wantbody: "{\"Items\":[{\"Balance\":",
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.req)
			res := w.Result()

			body, _ := ioutil.ReadAll(res.Body)

			if res.StatusCode != tt.status {
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
				return
			}

			if !strings.Contains(string(body), tt.wantbody) {
				t.Errorf("expected err to contain '%s', was '%s'", tt.wantbody, body)
			}
		})
	}
}

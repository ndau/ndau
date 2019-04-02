package routes_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
)

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
		name     string
		req      *http.Request
		status   int
		wantbody string
	}{
		{
			name:     "no address",
			req:      httptest.NewRequest("GET", "/account/history/", nil),
			status:   http.StatusBadRequest,
			wantbody: "address parameter required",
		}, {
			name:     "invalid address",
			req:      httptest.NewRequest("GET", "/account/history/"+invalidAddr, nil),
			status:   http.StatusBadRequest,
			wantbody: "could not validate address",
		}, {
			name:   "valid hash",
			req:    httptest.NewRequest("GET", "/account/history/"+addr, nil),
			status: http.StatusOK,
			// The addr isn't part of the response, just make sure the response is non-empty.
			wantbody: "{\"Items\":[{\"Balance\":",
		}, {
			name: "invalid page index",
			req: httptest.NewRequest("GET",
				fmt.Sprintf("/account/history/%s?pageindex=not_a_number", addr), nil),
			status:   http.StatusBadRequest,
			wantbody: "pageindex must be a valid number",
		}, {
			name: "invalid page size",
			req: httptest.NewRequest("GET",
				fmt.Sprintf("/account/history/%s?pagesize=not_a_number", addr), nil),
			status:   http.StatusBadRequest,
			wantbody: "pagesize must be a valid number",
		}, {
			name: "invalid page size number",
			req: httptest.NewRequest("GET",
				fmt.Sprintf("/account/history/%s?pagesize=%d", addr, -3), nil),
			status:   http.StatusBadRequest,
			wantbody: "pagesize must be non-negative",
		}, {
			name: "valid page",
			req: httptest.NewRequest("GET",
				fmt.Sprintf("/account/history/%s?pageindex=%d&pagesize=%d", addr, 0, 1), nil),
			status: http.StatusOK,
			// We know the first transaction will have zero balance since it's an account claim.
			wantbody: "{\"Items\":[{\"Balance\":0,\"Timestamp\":",
		}, {
			name: "valid end page",
			req: httptest.NewRequest("GET",
				fmt.Sprintf("/account/history/%s?pageindex=%d&pagesize=%d", addr, -1, 1), nil),
			status: http.StatusOK,
			// We don't know what the end balance will be, so we just check the response start.
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

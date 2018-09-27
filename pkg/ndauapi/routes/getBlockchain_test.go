package routes_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func TestBlockchain(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("early exit for integration tests")
	}

	baseHandler := routes.GetBlockchain

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "empty request",
			req:    httptest.NewRequest("GET", "/", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "missing start request",
			req:    httptest.NewRequest("GET", "/?end=2", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/?start=one&end=2", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "start too low request",
			req:    httptest.NewRequest("GET", "/?start=0&end=2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "end too low request",
			req:    httptest.NewRequest("GET", "/?start=1&end=0", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start higher than end request",
			req:    httptest.NewRequest("GET", "/?start=4&end=2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "range greater than 100 request",
			req:    httptest.NewRequest("GET", "/?start=1&end=102", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end == 3",
			req:    httptest.NewRequest("GET", "/?start=3&end=3", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET", "/?start=1&end=2", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, _ := cfg.New()
	handler := baseHandler(cf)

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler(w, tt.req)
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
		badHandler(w, httptest.NewRequest("GET", "/?start=50&end=51", nil))
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			body, _ := ioutil.ReadAll(res.Body)
			t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, http.StatusInternalServerError, body)
		}
	})
}

package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func TestNumUnconfirmedTxs(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("early exit for integration tests")
	}

	baseHandler := routes.GetNumUnconfirmedTxs

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "good request",
			req:    httptest.NewRequest("GET", "/", nil),
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
				t.Errorf("got status code %v, want %v", res.StatusCode, tt.status)
			}
		})
	}

	badHandler := baseHandler(cfg.Cfg{})
	t.Run("empty config", func(t *testing.T) {
		w := httptest.NewRecorder()
		badHandler(w, httptest.NewRequest("GET", "/", nil))
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("got status code %v, want %v", res.StatusCode, http.StatusInternalServerError)
		}
	})
}

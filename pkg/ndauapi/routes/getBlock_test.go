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

func TestBlock(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("early exit for integration tests")
	}

	baseHandler := routes.GetBlock

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "good request",
			req:    httptest.NewRequest("GET", "/?height=1", nil),
			status: http.StatusOK,
		},
		{
			name:   "right type, bad height request",
			req:    httptest.NewRequest("GET", "/?height=0", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "bad height request",
			req:    httptest.NewRequest("GET", "/?height=high", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "empty height request",
			req:    httptest.NewRequest("GET", "/?height=", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "no height request",
			req:    httptest.NewRequest("GET", "/", nil),
			status: http.StatusBadRequest,
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
		badHandler(w, httptest.NewRequest("GET", "/?height=1", nil))
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("got status code %v, want %v", res.StatusCode, http.StatusInternalServerError)
		}
	})
}

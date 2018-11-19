package routes_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
)

func TestNetInfo(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	baseHandler := routes.GetNetInfo

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
		badHandler(w, httptest.NewRequest("GET", "/", nil))
		res := w.Result()
		if res.StatusCode != http.StatusInternalServerError {
			t.Errorf("got status code %v, want %v", res.StatusCode, http.StatusInternalServerError)
		}
	})
}

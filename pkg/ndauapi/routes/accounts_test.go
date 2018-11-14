package routes_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
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
			status: http.StatusInternalServerError,
		},
	}

	// set up apparatus
	cf, _, _ := cfg.New()
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

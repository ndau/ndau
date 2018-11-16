package routes_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
)

func TestBlockHeight(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "good request",
			req:    httptest.NewRequest("GET", "/block/height/1", nil),
			status: http.StatusOK,
		},
		{
			name:   "right type, bad height request",
			req:    httptest.NewRequest("GET", "/block/height/99999999", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "bad height request",
			req:    httptest.NewRequest("GET", "/block/height/high", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "current height request",
			req:    httptest.NewRequest("GET", "/block/current", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, _ := cfg.New()
	mux := svc.New(cf).Mux()

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

}

func TestBlockRange(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/block/range/one/2", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "start too low request",
			req:    httptest.NewRequest("GET", "/block/range/0/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "end too low request",
			req:    httptest.NewRequest("GET", "/block/range/1/0", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start higher than end request",
			req:    httptest.NewRequest("GET", "/block/range/4/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "range greater than 100 request",
			req:    httptest.NewRequest("GET", "/block/range/1/102", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end == 3",
			req:    httptest.NewRequest("GET", "/block/range/3/3", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET", "/block/range/1/2", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, _ := cfg.New()
	mux := svc.New(cf).Mux()

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
}

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
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
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

	// If a reset has occurred recently, the blockchain height can sometimes be as low as 2.
	// The tests below require a height of at least 3.  Assume the height is zero, worst case.
	for i := 0; i < 3; i++ {
		createNdauBlock(t)
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
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
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

func TestChaosBlockRange(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// If a reset has occurred recently, the blockchain height can sometimes be as low as 2.
	// The tests below require a height of at least 3.  Assume the height is zero, worst case.
	for i := 0; i < 3; i++ {
		createChaosBlock(t, i)
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/chaos/range/one/2", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "start too low request",
			req:    httptest.NewRequest("GET", "/chaos/range/0/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "end too low request",
			req:    httptest.NewRequest("GET", "/chaos/range/1/0", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start higher than end request",
			req:    httptest.NewRequest("GET", "/chaos/range/4/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "range greater than 100 request",
			req:    httptest.NewRequest("GET", "/chaos/range/1/102", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end == 3",
			req:    httptest.NewRequest("GET", "/chaos/range/3/3", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET", "/chaos/range/1/2", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
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

func TestBlockDateRange(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// If a reset has occurred recently, the blockchain height can sometimes be as low as 2.
	// The tests below require a height of at least 3.  Assume the height is zero, worst case.
	for i := 0; i < 3; i++ {
		createNdauBlock(t)
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/block/daterange/one/2018-07-10T20:01:02Z", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end the same",
			req:    httptest.NewRequest("GET",
				"/block/daterange/2018-07-10T20:01:02Z/2018-07-10T20:01:02Z", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET",
				"/block/daterange/2018-07-10T00:00:00Z/2018-07-11T00:00:00Z", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
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

func TestChaosBlockDateRange(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// If a reset has occurred recently, the blockchain height can sometimes be as low as 2.
	// The tests below require a height of at least 3.  Assume the height is zero, worst case.
	for i := 0; i < 3; i++ {
		createChaosBlock(t, i)
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/chaos/daterange/one/2018-07-10T20:01:02Z", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end the same",
			req:    httptest.NewRequest("GET",
				"/chaos/daterange/2018-07-10T20:01:02Z/2018-07-10T20:01:02Z", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET",
				"/chaos/daterange/2018-07-10T00:00:00Z/2018-07-11T00:00:00Z", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
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

func TestBlockHash(t *testing.T) {
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
	createNdauBlock(t)

	// Grab the block hash for use in later tests.
	blockData := getCurrentNdauBlock(t, mux)
	blockHash := fmt.Sprintf("%x", blockData.BlockMeta.BlockID.Hash)

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "no hash",
			req:      httptest.NewRequest("GET", "/block/hash/", nil),
			status:   http.StatusBadRequest,
			wantbody: "blockhash parameter required",
		}, {
			name:     "invalid hash",
			req:      httptest.NewRequest("GET", "/block/hash/invalidhash", nil),
			status:   http.StatusOK,
			wantbody: "null", // The response is empty, so "null" is produced.
		}, {
			name:     "valid hash",
			req:      httptest.NewRequest("GET", "/block/hash/" + blockHash, nil),
			status:   http.StatusOK,
			wantbody: blockHash, // The response should contain the hash we searched for.
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
				t.Errorf("SubmitTx() expected err to contain '%s', was '%s'", tt.wantbody, body)
			}
		})
	}
}

func TestBlockTransactions(t *testing.T) {
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
	createNdauBlock(t)

	// Grab the block hash for use in later tests.
	blockData := getCurrentNdauBlock(t, mux)
	blockHeight := blockData.BlockMeta.Header.Height

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "good request",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/block/transactions/%d", blockHeight), nil),
			status:   http.StatusOK,
			wantbody: "[\"", // The response should contain a tx hash array.
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
				t.Errorf("SubmitTx() expected err to contain '%s', was '%s'", tt.wantbody, body)
			}
		})
	}
}

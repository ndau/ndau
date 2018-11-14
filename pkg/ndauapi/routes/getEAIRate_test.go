package routes_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func TestGetEAIRate(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	baseHandler := routes.GetEAIRate

	tests := []struct {
		name   string
		body   []routes.EAIRateRequest
		status int
		want   []routes.EAIRateResponse
	}{
		{
			name:   "empty request",
			body:   []routes.EAIRateRequest{},
			status: http.StatusOK,
			want:   []routes.EAIRateResponse{},
		},
		{
			name:   "no body",
			body:   nil,
			status: http.StatusBadRequest,
			want:   nil,
		},
		{
			name: "zero rate",
			body: []routes.EAIRateRequest{
				routes.EAIRateRequest{"zero", types.Duration(0), backing.Lock{}},
			},
			status: http.StatusOK,
			want: []routes.EAIRateResponse{
				routes.EAIRateResponse{"zero", 0},
			},
		},
		{
			name: "unlocked 3month rate",
			body: []routes.EAIRateRequest{
				routes.EAIRateRequest{"3L0", types.Month * 3, backing.Lock{}},
			},
			status: http.StatusOK,
			want: []routes.EAIRateResponse{
				routes.EAIRateResponse{"3L0", uint64(eai.RateFromPercent(4))},
			},
		},
		{
			name: "locked 90 days at time 0",
			body: []routes.EAIRateRequest{
				routes.EAIRateRequest{"0L90", 0, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
			},
			status: http.StatusOK,
			want: []routes.EAIRateResponse{
				routes.EAIRateResponse{"0L90", uint64(eai.RateFromPercent(1))},
			},
		},
		{
			name: "several accounts",
			body: []routes.EAIRateRequest{
				routes.EAIRateRequest{"90L90", 90 * types.Day, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
				routes.EAIRateRequest{"0L90", 0, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
				routes.EAIRateRequest{"90L0", 90 * types.Day, backing.Lock{}},
				routes.EAIRateRequest{"400L1095", 400 * types.Day, *backing.NewLock(1095*types.Day, eai.DefaultLockBonusEAI)},
			},
			status: http.StatusOK,
			want: []routes.EAIRateResponse{
				routes.EAIRateResponse{"90L90", uint64(eai.RateFromPercent(5))},
				routes.EAIRateResponse{"0L90", uint64(eai.RateFromPercent(1))},
				routes.EAIRateResponse{"90L0", uint64(eai.RateFromPercent(4))},
				routes.EAIRateResponse{"400L1095", uint64(eai.RateFromPercent(15))},
			},
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
			var req *http.Request
			if tt.body != nil {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(tt.body)
				req = httptest.NewRequest("GET", "/", buf)
			} else {
				req = httptest.NewRequest("GET", "/", nil)
			}
			handler(w, req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
				return
			}
			var got []routes.EAIRateResponse
			json.NewDecoder(res.Body).Decode(&got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEAIRate() = %v, want %v", got, tt.want)
			}
		})
	}

}

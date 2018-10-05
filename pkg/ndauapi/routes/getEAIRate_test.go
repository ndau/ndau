package routes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/oneiro-ndev/ndaumath/pkg/eai"

	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
)

func TestGetEAIRate(t *testing.T) {
	baseHandler := GetEAIRate

	tests := []struct {
		name   string
		body   []EAIRateRequest
		status int
		want   []EAIRateResponse
	}{
		{
			name:   "empty request",
			body:   []EAIRateRequest{},
			status: http.StatusOK,
			want:   []EAIRateResponse{},
		},
		{
			name:   "no body",
			body:   nil,
			status: http.StatusBadRequest,
			want:   nil,
		},
		{
			name: "zero rate",
			body: []EAIRateRequest{
				EAIRateRequest{"zero", types.Duration(0), backing.Lock{}},
			},
			status: http.StatusOK,
			want: []EAIRateResponse{
				EAIRateResponse{"zero", 0},
			},
		},
		{
			name: "unlocked 3month rate",
			body: []EAIRateRequest{
				EAIRateRequest{"3L0", types.Month * 3, backing.Lock{}},
			},
			status: http.StatusOK,
			want: []EAIRateResponse{
				EAIRateResponse{"3L0", 4000000},
			},
		},
		{
			name: "locked 90 days at time 0",
			body: []EAIRateRequest{
				EAIRateRequest{"0L90", 0, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
			},
			status: http.StatusOK,
			want: []EAIRateResponse{
				EAIRateResponse{"0L90", 1000000},
			},
		},
		{
			name: "several accounts",
			body: []EAIRateRequest{
				EAIRateRequest{"90L90", 90 * types.Day, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
				EAIRateRequest{"0L90", 0, *backing.NewLock(90*types.Day, eai.DefaultLockBonusEAI)},
				EAIRateRequest{"90L0", 90 * types.Day, backing.Lock{}},
				EAIRateRequest{"400L1095", 400 * types.Day, *backing.NewLock(1095*types.Day, eai.DefaultLockBonusEAI)},
			},
			status: http.StatusOK,
			want: []EAIRateResponse{
				EAIRateResponse{"90L90", 5000000},
				EAIRateResponse{"0L90", 1000000},
				EAIRateResponse{"90L0", 4000000},
				EAIRateResponse{"400L1095", 15000000},
			},
		},
	}

	// set up apparatus
	cf, _, _ := cfg.New()
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
			var got []EAIRateResponse
			json.NewDecoder(res.Body).Decode(&got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEAIRate() = %v, want %v", got, tt.want)
			}
		})
	}

}

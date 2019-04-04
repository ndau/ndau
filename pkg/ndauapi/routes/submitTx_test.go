package routes_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSubmitTxNoServer(t *testing.T) {
	keypub, keypvt, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindUser, keypub.KeyBytes())
	require.NoError(t, err)
	testLockTx := ndau.NewLock(addr, 30*types.Day, 1234, keypvt)

	tests := []struct {
		name    string
		body    interface{}
		status  int
		want    routes.SubmitResult
		wanterr string
	}{
		{
			name:    "no body",
			body:    nil,
			status:  http.StatusBadRequest,
			want:    routes.SubmitResult{},
			wanterr: "did not unmarshal",
		},
		{
			name:    "valid tx but no node",
			body:    testLockTx,
			status:  http.StatusInternalServerError,
			want:    routes.SubmitResult{},
			wanterr: "error retrieving node",
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	require.Error(t, err)

	mux := svc.New(cf).Mux()

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			var req *http.Request
			if tt.body != nil {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(tt.body)
				req = httptest.NewRequest("POST", "/tx/submit/lock", buf)
			} else {
				req = httptest.NewRequest("POST", "/tx/submit/lock", nil)
			}
			mux.ServeHTTP(w, req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
				return
			}

			var got routes.SubmitResult
			err := json.NewDecoder(res.Body).Decode(&got)
			if err != nil {
				t.Errorf("Error decoding result: %s", err)
			}
			if tt.wanterr != "" {
				if !strings.Contains(got.Msg, tt.wanterr) {
					t.Errorf("SubmitTx() expected err to contain %s, was %s", tt.wanterr, got.Msg)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SubmitTx() = %v, want %v", got, tt.want)
			}
		})
	}
}

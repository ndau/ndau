package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestPrevalidateTxNoServer(t *testing.T) {
	keypub, keypvt, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindUser, keypub.KeyBytes())
	require.NoError(t, err)
	testLockTx := ndau.NewLock(addr, 30*types.Day, 1234)
	testLockTx.Signatures = append(testLockTx.Signatures, metatx.Sign(testLockTx, keypvt))
	require.NoError(t, err)

	tests := []struct {
		name    string
		body    interface{}
		status  int
		want    routes.SubmitResult
		wanterr string
		skip    bool
	}{
		{
			name:    "no body",
			body:    nil,
			status:  http.StatusBadRequest,
			want:    routes.SubmitResult{},
			wanterr: "did not unmarshal",
			skip:    false,
		},
		{
			name:    "valid tx but no node",
			body:    testLockTx,
			status:  http.StatusInternalServerError,
			want:    routes.SubmitResult{},
			wanterr: "error retrieving node",
			skip:    isIntegration,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	// Integration tests must have a valid config, non-integration tests expect an invalid one.
	if isIntegration == (err != nil) {
		t.Errorf("Unexpected config error: %s", err)
		return
	}
	mux := svc.New(cf).Mux()

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(fmt.Sprintf("isIntegration is %v", isIntegration))
			}

			w := httptest.NewRecorder()
			var req *http.Request
			if tt.body != nil {
				buf := &bytes.Buffer{}
				json.NewEncoder(buf).Encode(tt.body)
				req = httptest.NewRequest("POST", "/tx/prevalidate/lock", buf)
			} else {
				req = httptest.NewRequest("POST", "/tx/prevalidate/lock", nil)
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
				return
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

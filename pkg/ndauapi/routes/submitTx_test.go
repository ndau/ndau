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

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSubmitTxNoServer(t *testing.T) {
	baseHandler := routes.HandleSubmitTx

	keypub, _, err := signature.Generate(signature.Ed25519, nil)
	require.NoError(t, err)
	addr, err := address.Generate(address.KindUser, keypub.KeyBytes())
	require.NoError(t, err)
	testLockTx := ndau.NewLock(addr, 30*types.Day, 1234, nil)
	testLockData, err := b64Tx(testLockTx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		body    *routes.TxJSON
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
			wanterr: "unable to decode",
			skip:    !isIntegration,
		},
		{
			name:    "blank request",
			body:    &routes.TxJSON{},
			status:  http.StatusBadRequest,
			want:    routes.SubmitResult{},
			wanterr: "could not be decoded",
			skip:    !isIntegration,
		},
		{
			name: "not base64",
			body: &routes.TxJSON{
				Data: "not base64 tx data",
			},
			status:  http.StatusBadRequest,
			want:    routes.SubmitResult{},
			wanterr: "could not be decoded as base64",
			skip:    !isIntegration,
		},
		{
			name: "not a tx",
			body: &routes.TxJSON{
				Data: b64str("not a tx"),
			},
			status:  http.StatusBadRequest,
			want:    routes.SubmitResult{},
			wanterr: "could not be decoded into a transaction",
			skip:    !isIntegration,
		},
		{
			name: "valid tx but no node",
			body: &routes.TxJSON{
				Data: testLockData,
			},
			status:  http.StatusInternalServerError,
			want:    routes.SubmitResult{},
			wanterr: "error retrieving node",
			skip:    isIntegration,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	// Integration tests must have a valid config, non-integration tests expect an invalid one.
	if err != nil && isIntegration {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
	handler := baseHandler(cf)

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
				req = httptest.NewRequest("POST", "/", buf)
			} else {
				req = httptest.NewRequest("POST", "/", nil)
			}
			handler(w, req)
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

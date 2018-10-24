package routes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
)

func TestTxClaimAccountNoServer(t *testing.T) {
	baseHandler := HandleClaimAccount

	ownerpub, _, _ := signature.Generate(signature.Ed25519, nil)
	addr, _ := address.Generate(address.KindUser, ownerpub.KeyBytes())
	valpub, _, _ := signature.Generate(signature.Ed25519, nil)

	tests := []struct {
		name   string
		body   interface{}
		status int
		want   PreparedTx
	}{
		{
			name: "valid tx",
			body: &TxClaimAccountRequest{
				Target:         addr,
				OwnershipKey:   ownerpub,
				ValidationKeys: []signature.PublicKey{valpub},
				Sequence:       12349,
			},
			status: http.StatusOK,
			want: PreparedTx{
				TxData:        "some encoded data",
				SignableBytes: "some more encoded data",
				Signatures:    nil,
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

			var got PreparedTx
			err := json.NewDecoder(res.Body).Decode(&got)
			if err != nil {
				t.Errorf("Error decoding result: %s", err)
			}
			// the tx will be different every time so we can't compare exactly
			if len(got.TxData) < 80 || len(got.SignableBytes) < 80 {
				t.Errorf("SubmitTx() = %#v, want %#v", got, tt.want)
			}
		})
	}

}

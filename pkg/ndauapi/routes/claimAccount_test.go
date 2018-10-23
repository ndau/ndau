package routes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
		name    string
		body    interface{}
		status  int
		want    PreparedTx
		wanterr string
	}{
		// {
		// 	name:   "no body",
		// 	body:   nil,
		// 	status: http.StatusBadRequest,
		// 	want:   PreparedTx{},
		// 	// wanterr: "unable to decode",
		// },
		// {
		// 	name:   "blank request",
		// 	body:   &TxClaimAccountRequest{},
		// 	status: http.StatusBadRequest,
		// 	want:   PreparedTx{},
		// 	// wanterr: "could not be decoded",
		// },
		// {
		// 	name: "not the proper type of JSON object",
		// 	body: &PreparedTx{
		// 		TxData:        "not base64 tx data",
		// 		SignableBytes: "not base64 bytes to be signed",
		// 		Signatures:    []string{"base64 signature of SignableBytes"},
		// 	},
		// 	status: http.StatusBadRequest,
		// 	want:   PreparedTx{},
		// 	// wanterr: "could not be decoded as base64",
		// },
		// {
		// 	name: "not a proper account",
		// 	body: `{
		//   "target": "ndaanqp9wz5jxgdynttx3chq98ach7a54hgnvfb2tdzdsmup",
		//   "ownership": "npuba8jadtbbecrs7duz6xz78fd9hz7tempfh4gxzkwwuniumfn2dbje7wxswf8kp3uji2dux7w3",
		//   "keys": [
		//     "npuba8jadtbbecrs7duz6xz78fd9hz7tempfh4gxzkwwuniumfn2dbje7wxswf8kp3uji2dux7w3"
		//   ],
		//   "script": "",
		//   "seq": 13579
		// }`,
		// 	status: http.StatusBadRequest,
		// 	want:   PreparedTx{},
		// 	// wanterr: "could not be decoded into a transaction",
		// },
		{
			name: "valid tx",
			body: &TxClaimAccountRequest{
				Target:         addr,
				OwnershipKey:   ownerpub,
				ValidationKeys: []signature.PublicKey{valpub},
				Sequence:       12349,
			},
			status: http.StatusOK,
			want:   PreparedTx{},
			// wanterr: "error retrieving node",
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

			var got TxResult
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

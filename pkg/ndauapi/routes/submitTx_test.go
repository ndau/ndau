package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndau"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/types"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
)

func TestSubmitTx(t *testing.T) {
	baseHandler := HandleSubmitTx

	b64 := func(b []byte) string {
		return base64.StdEncoding.EncodeToString(b)
	}

	b64str := func(s string) string {
		return b64([]byte(s))
	}

	b64Tx := func(tx ndau.NTransactable) string {
		var b []byte
		m, err := tx.MarshalMsg(b)
		if err != nil {
			panic("marshal failed!!!")
		}
		return b64(m)
	}

	addr, _ := address.Generate(address.KindUser, []byte("somerandomdatatomakeaddressfrom"))
	testLockTx := ndau.NewLock(addr, 30*types.Day, 1234, nil)

	tests := []struct {
		name    string
		body    *PreparedTx
		status  int
		want    TxResult
		wanterr string
	}{
		{
			name:    "no body",
			body:    nil,
			status:  http.StatusBadRequest,
			want:    TxResult{},
			wanterr: "unable to decode",
		},
		{
			name:    "blank request",
			body:    &PreparedTx{},
			status:  http.StatusBadRequest,
			want:    TxResult{},
			wanterr: "could not be decoded",
		},
		{
			name: "not base64",
			body: &PreparedTx{
				TxData:        "not base64 tx data",
				SignableBytes: "not base64 bytes to be signed",
				Signatures:    []string{"base64 signature of SignableBytes"},
			},
			status:  http.StatusBadRequest,
			want:    TxResult{},
			wanterr: "could not be decoded as base64",
		},
		{
			name: "not a tx",
			body: &PreparedTx{
				TxData:        b64str("not a tx"),
				SignableBytes: b64str("not signable"),
				Signatures:    []string{b64str("not a signature of SignableBytes")},
			},
			status:  http.StatusBadRequest,
			want:    TxResult{},
			wanterr: "could not be decoded into a transaction",
		},
		{
			name: "valid tx",
			body: &PreparedTx{
				TxData:        b64Tx(testLockTx),
				SignableBytes: b64(testLockTx.SignableBytes()),
				Signatures:    []string{},
			},
			status: http.StatusBadRequest,
			want:   TxResult{},
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

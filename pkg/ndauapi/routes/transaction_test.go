package routes_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	metatx "github.com/oneiro-ndev/metanode/pkg/meta/transaction"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
)

func TestTxHash(t *testing.T) {
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
	txBytes := blockData.Block.Data.Txs[0] // Expecting one transaction, get the first.
	txAble, err := metatx.Unmarshal(txBytes, ndau.TxIDs)
	if err != nil {
		t.Errorf("Error unmarshaling tx: %s", err)
	}
	txHash := metatx.Hash(txAble)
	txHashEsc := url.QueryEscape(txHash)

	// Invalid in that it's not valid base64, and it's not in the index.
	invalidTxHash := "invalid=hash_"
	invalidTxHashEsc := url.QueryEscape(invalidTxHash)

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "no hash",
			req:      httptest.NewRequest("GET", "/transaction/", nil),
			status:   http.StatusBadRequest,
			wantbody: "txhash parameter required",
		}, {
			name:     "invalid hash",
			req:      httptest.NewRequest("GET", "/transaction/" + invalidTxHashEsc, nil),
			status:   http.StatusOK,
			wantbody: "null", // The response is empty, so "null" is produced.
		}, {
			name:     "valid hash",
			req:      httptest.NewRequest("GET", "/transaction/" + txHashEsc, nil),
			status:   http.StatusOK,
			// The tx hash isn't part of the response, just make sure a valid tx is returned.
			wantbody: "{\"Tx\":{\"Nonce\":",
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
				t.Errorf("expected err to contain '%s', was '%s'", tt.wantbody, body)
			}
		})
	}
}

package routes_test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
)

func TestKeyHistory(t *testing.T) {
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

	namespaceBase64, key, value := createChaosBlock(t)
	invalidNamespace := "invalidnamespace"
	invalidKey := "invalidkey"

	// Namespaces are base64, but keys aren't.  Both must be path escaped to pass on the URL.
	namespaceBase64Esc := url.PathEscape(namespaceBase64)
	invalidNamespaceBase64 := base64.StdEncoding.EncodeToString([]byte(invalidNamespace))
	invalidNamespaceBase64Esc := url.PathEscape(invalidNamespaceBase64)
	keyEsc := url.PathEscape(key)
	invalidKeyEsc := url.PathEscape(invalidKey)
	valueBase64 := base64.StdEncoding.EncodeToString([]byte(value))

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "no namespace no key",
			req:      httptest.NewRequest("GET", "/chaos/history/", nil),
			status:   http.StatusFound,
			wantbody: "", // Blank response, need namespace and key or it won't get very far.
		}, {
			name:     "no key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s", namespaceBase64Esc), nil),
			status:   http.StatusNotFound,
			wantbody: "page not found", // Need namespace and key or it won't find the handler.
		}, {
			name:     "invalid namespace and key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", invalidNamespaceBase64Esc, invalidKeyEsc), nil),
			status:   http.StatusOK,
			wantbody: "{\"History\":null}", // Successful but empty response.
		}, {
			name:     "valid namespace and key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", namespaceBase64Esc, keyEsc), nil),
			status:   http.StatusOK,
			// The key isn't part of the response, just make sure a non-empty history is returned.
			wantbody: fmt.Sprintf("\"Value\":\"%s\"", valueBase64),
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

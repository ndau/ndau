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

	namespaceBase64, key, value := createChaosBlock(t, 0)

	// Use underscores so they're not valid base64, in addition to not being indexed.
	invalidNamespace := "invalid=namespace_"
	invalidKey := "invalid=key_"

	// Namespaces and keys are base64.  Both must be query escaped to pass on the URL.
	namespaceBase64Esc := url.QueryEscape(namespaceBase64)
	invalidNamespaceBase64 := base64.StdEncoding.EncodeToString([]byte(invalidNamespace))
	invalidNamespaceBase64Esc := url.QueryEscape(invalidNamespaceBase64)
	keyBase64 := base64.StdEncoding.EncodeToString([]byte(key))
	keyBase64Esc := url.QueryEscape(keyBase64)
	invalidKeyBase64 := base64.StdEncoding.EncodeToString([]byte(invalidKey))
	invalidKeyBase64Esc := url.QueryEscape(invalidKeyBase64)
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
			name:     "invalid base64 namespace and base64 key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", invalidNamespace, invalidKey), nil),
			status:   http.StatusBadRequest,
			wantbody: "error decoding namespace",
		}, {
			name:     "invalid base64 namespace and key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", invalidNamespace, invalidKeyBase64Esc), nil),
			status:   http.StatusBadRequest,
			wantbody: "error decoding namespace",
		}, {
			name:     "invalid namespace and base64 key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", invalidNamespaceBase64Esc, invalidKey), nil),
			status:   http.StatusBadRequest,
			wantbody: "error decoding key",
		}, {
			name:     "invalid namespace and key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", invalidNamespaceBase64Esc, invalidKeyBase64Esc), nil),
			status:   http.StatusOK,
			wantbody: "{\"History\":null}", // Successful but empty response.
		}, {
			name:     "valid namespace and key",
			req:      httptest.NewRequest("GET", fmt.Sprintf("/chaos/history/%s/%s", namespaceBase64Esc, keyBase64Esc), nil),
			status:   http.StatusOK,
			// The key isn't part of the response, but we can look for the expected value.
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

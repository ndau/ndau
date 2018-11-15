package routes_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func TestBlockHeight(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "good request",
			req:    httptest.NewRequest("GET", "/block/height/1", nil),
			status: http.StatusOK,
		},
		{
			name:   "right type, bad height request",
			req:    httptest.NewRequest("GET", "/block/height/99999999", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "bad height request",
			req:    httptest.NewRequest("GET", "/block/height/high", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "current height request",
			req:    httptest.NewRequest("GET", "/block/current", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
	mux := svc.New(cf).Mux()

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
			}
		})
	}

}

func TestBlockRange(t *testing.T) {
	if !isIntegration {
		t.Skip("integration tests are opt-in")
	}

	// set up tests
	tests := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{
			name:   "start bad number request",
			req:    httptest.NewRequest("GET", "/block/range/one/2", nil),
			status: http.StatusBadRequest,
		},
		{
			name:   "start too low request",
			req:    httptest.NewRequest("GET", "/block/range/0/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "end too low request",
			req:    httptest.NewRequest("GET", "/block/range/1/0", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start higher than end request",
			req:    httptest.NewRequest("GET", "/block/range/4/2", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "range greater than 100 request",
			req:    httptest.NewRequest("GET", "/block/range/1/102", nil),
			status: http.StatusBadRequest,
		}, {
			name:   "start and end == 3",
			req:    httptest.NewRequest("GET", "/block/range/3/3", nil),
			status: http.StatusOK,
		}, {
			name:   "good request",
			req:    httptest.NewRequest("GET", "/block/range/1/2", nil),
			status: http.StatusOK,
		},
	}

	// set up apparatus
	cf, _, err := cfg.New()
	if err != nil {
		t.Errorf("Error creating cfg: %s", err)
		return
	}
	mux := svc.New(cf).Mux()

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, tt.req)
			res := w.Result()
			if res.StatusCode != tt.status {
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("got status code %v, want %v. (%s)", res.StatusCode, tt.status, body)
			}
		})
	}
}

func TestBlockHash(t *testing.T) {
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

	// Add new blocks that get indexed so we can search on them.
	goOutput, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		t.Errorf("Error getting go dir: %s", err)
	}

	goDir := bytes.TrimRight(goOutput, "\n")
	ndauTool := fmt.Sprintf("%s/src/github.com/oneiro-ndev/ndau/ndau", goDir)
	acctName := "integrationtest"

	acctCmd := exec.Command(ndauTool, "account", "new", acctName)
	acctStderr, err := acctCmd.StderrPipe()
	if err != nil {
		t.Errorf("Error creating account command: %s", err)
	}
	err = acctCmd.Start(); 
	if err != nil {
		t.Errorf("Error starting account command: %s", err)
	}
	acctErr, _ := ioutil.ReadAll(acctStderr)
	err = acctCmd.Wait()
	if err == nil {
		err = exec.Command(ndauTool, "account", "claim", acctName).Run()
		if err != nil {
			t.Errorf("Error claiming account: %s", err)
		}
	} else if !strings.Contains(string(acctErr), "already exists") {
		t.Errorf("Error creating account: %s", err)
	}

	err = exec.Command(ndauTool, "-v", "rfe", "10", acctName).Run()
	if err != nil {
		t.Errorf("Error claiming account: %s", err)
	}

	// We will generate a tx, then get its block hash to search for.
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/block/current", nil)
	mux.ServeHTTP(w, req)
	res := w.Result()

	var got rpctypes.ResultBlock
	err = json.NewDecoder(res.Body).Decode(&got)
	if err != nil {
		t.Errorf("Error decoding result: %s", err)
	}

	// Grab the block hash for use in later tests.
	blockHash := fmt.Sprintf("%x", got.BlockMeta.BlockID.Hash)

	// set up tests
	tests := []struct {
		name   	 string
		req    	 *http.Request
		status 	 int
		wantbody string
	}{
		{
			name:     "no hash",
			req:      httptest.NewRequest("GET", "/block/hash/", nil),
			status:   http.StatusBadRequest,
			wantbody: "blockhash parameter required",
		}, {
			name:     "invalid hash",
			req:      httptest.NewRequest("GET", "/block/hash/invalidhash", nil),
			status:   http.StatusOK,
			wantbody: "",
		}, {
			name:     "valid hash",
			req:      httptest.NewRequest("GET", "/block/hash/" + blockHash, nil),
			status:   http.StatusOK,
			wantbody: blockHash,
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
			}

			if !strings.Contains(string(body), tt.wantbody) {
				t.Errorf("SubmitTx() expected err to contain '%s', was '%s'", tt.wantbody, body)
			}
		})
	}
}

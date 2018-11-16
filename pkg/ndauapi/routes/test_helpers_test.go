package routes_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var isIntegration bool
var ndauRPC string
var chaosRPC string

func init() {
	flag.BoolVar(&isIntegration, "integration", false, "opt into integration tests")
	flag.StringVar(&ndauRPC, "ndaurpc", "", "ndau rpc url")
	flag.StringVar(&chaosRPC, "chaosrpc", "", "chaos rpc url")
	flag.Parse()

	// Put these into environment variables so cfg.New() finds them where it's looking for them.
	// But only if they were specified on the command line.  Otherwise, they may already be in
	// environment variables.
	if ndauRPC != "" {
		os.Setenv("NDAUAPI_NDAU_RPC_URL", ndauRPC)
	}
	if chaosRPC != "" {
		os.Setenv("NDAUAPI_CHAOS_RPC_URL", chaosRPC)
	}
}

// Invoke the ndau tool to create an account, claim it, and rfe to it.
func createBlock(t *testing.T) {
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
}

// Use the /block/current endpoint to grab the head block.
func getCurrentBlock(t *testing.T, mux http.Handler) (blockData rpctypes.ResultBlock) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/block/current", nil)
	mux.ServeHTTP(w, req)
	res := w.Result()

	err := json.NewDecoder(res.Body).Decode(&blockData)
	if err != nil {
		t.Errorf("Error decoding result: %s", err)
	}

	return
}

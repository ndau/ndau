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

// Return the system's go directory.
func getGoDir(t *testing.T) string {
	output, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		t.Errorf("Error getting go dir: %s", err)
	}

	return string(bytes.TrimRight(output, "\n"))
}

// Invoke the ndau tool to create an account, claim it, and rfe to it.
func createNdauBlock(t *testing.T) {
	acctName := "integrationtestacct"

	goDir := getGoDir(t)
	ndauTool := fmt.Sprintf("%s/src/github.com/oneiro-ndev/ndau/ndau", goDir)

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
		t.Errorf("Error issuing ndau: %s", err)
	}
}

// Use the /block/current endpoint to grab the head block.
func getCurrentNdauBlock(t *testing.T, mux http.Handler) (blockData rpctypes.ResultBlock) {
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

// Invoke the chaos tool to create a key value pair in the chaos chain.
// Return the namespace, key and value that were set.
func createChaosBlock(t *testing.T) (namespaceBase64, key, value string) {
	namespace := "integrationtestnamespace"
	key = "integrationtestkey"
	value = "integrationtestvalue"

	goDir := getGoDir(t)
	chaosTool := fmt.Sprintf("%s/src/github.com/oneiro-ndev/chaos/chaos", goDir)

	// Create a test namespace.
	namespaceCmd := exec.Command(chaosTool, "id", "new", namespace)
	namespaceStderr, err := namespaceCmd.StderrPipe()
	if err != nil {
		t.Errorf("Error creating namespace command: %s", err)
	}
	err = namespaceCmd.Start(); 
	if err != nil {
		t.Errorf("Error starting namespace command: %s", err)
	}
	namespaceErr, _ := ioutil.ReadAll(namespaceStderr)
	err = namespaceCmd.Wait()
	if err != nil && !strings.Contains(string(namespaceErr), "already present") {
		t.Errorf("Error creating namespace: %s", err)
	}

	// Set a k-v pair.
	err = exec.Command(chaosTool, "set", namespace, "-k=" + key, "-v=" + value).Run()
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}

	// Get the binary form of the namespace we created.
	listCmd := exec.Command(chaosTool, "id", "list")
	listStdout, err := listCmd.StdoutPipe()
	if err != nil {
		t.Errorf("Error creating list command: %s", err)
	}
	err = listCmd.Start(); 
	if err != nil {
		t.Errorf("Error starting list command: %s", err)
	}
	listOutput, _ := ioutil.ReadAll(listStdout)
	err = listCmd.Wait()
	if err != nil {
		t.Errorf("Error getting namespace list: %s", err)
	}

	// Parse the output.
	namespaceBase64 = ""
	lines := strings.Split(string(listOutput), "\n")
	for _, line := range lines {
		firstSpace := strings.Index(line, " ")
		if firstSpace > 0 {
			ns := line[:firstSpace]
			if ns == namespace {
				lastSpace := strings.LastIndex(line, " ")
				namespaceBase64 = line[lastSpace+1:]
				break
			}
		}
	}
	if namespaceBase64 == "" {
		t.Error("Unable to deduce base64 namespace")
	}

	return
}

package routes_test

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// This is used as an ndau account name as well as a chaos namespace in our tests.
const accountAndNamespaceString = "integrationtest"

func init() {
	// The config calls need the NDAUHOME env var set, as well as each chaos/ndau tool invocation.
	ndauhome := os.ExpandEnv("$NDAUHOME")
	if ndauhome == "" {
		userhome := os.ExpandEnv("$HOME")
		// Every localnet, multi-node or not, has a 0'th node we can use.
		ndauhome = filepath.Join(userhome, ".localnet/data/ndau-0")
		os.Setenv("NDAUHOME", ndauhome)
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

// Invoke the ndau tool to create an account, set its validation rules, and rfe to it.
// Return the address of the account created.
func createNdauBlock(t *testing.T) string {
	acctName := accountAndNamespaceString

	goDir := getGoDir(t)
	ndauTool := fmt.Sprintf("%s/src/github.com/oneiro-ndev/commands/ndau", goDir)

	acctCmd := exec.Command(ndauTool, "account", "new", acctName)
	acctStderr, err := acctCmd.StderrPipe()
	if err != nil {
		t.Errorf("Error creating account command: %s", err)
	}
	err = acctCmd.Start()
	if err != nil {
		t.Errorf("Error starting account command: %s", err)
	}
	acctErr, _ := ioutil.ReadAll(acctStderr)
	err = acctCmd.Wait()
	if err == nil {
		err = exec.Command(ndauTool, "account", "set-validation", acctName).Run()
		if err != nil {
			t.Errorf("Error setting validation rules for account: %s", err)
		}
	} else if !strings.Contains(string(acctErr), "already exists") {
		t.Errorf("Error creating account: %s", err)
	}

	err = exec.Command(ndauTool, "rfe", "10", acctName).Run()
	if err != nil {
		t.Errorf("Error issuing ndau: %s", err)
	}

	c, err := config.LoadDefault(config.GetConfigPath())
	if err != nil {
		t.Errorf("Error getting config: %s", err)
	}

	addr := ""
	for _, acct := range c.GetAccounts() {
		if acct.Name == acctName {
			addr = acct.Address.String()
			break
		}
	}
	if addr == "" {
		t.Error("Unable to get account address")
	}

	return addr
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
func createChaosBlock(t *testing.T, valnum int) (namespaceBase64, key, value string) {
	namespace := accountAndNamespaceString
	key = "integrationtestkey"
	value = fmt.Sprintf("integrationtestvalue%d", valnum)

	goDir := getGoDir(t)
	chaosTool := fmt.Sprintf("%s/src/github.com/oneiro-ndev/commands/chaos", goDir)

	// Create a test namespace.
	namespaceCmd := exec.Command(chaosTool, "id", "new", namespace)
	namespaceStderr, err := namespaceCmd.StderrPipe()
	if err != nil {
		t.Errorf("Error creating namespace command: %s", err)
	}
	err = namespaceCmd.Start()
	if err != nil {
		t.Errorf("Error starting namespace command: %s", err)
	}
	namespaceErr, _ := ioutil.ReadAll(namespaceStderr)
	err = namespaceCmd.Wait()
	if err != nil && !strings.Contains(string(namespaceErr), "already present") {
		t.Errorf("Error creating namespace: %s", err)
	}

	// This creates the account from which our chaos transactions will pull ndau from for tx fees.
	createNdauBlock(t)

	// Associate the namespace with the account.
	err = exec.Command(chaosTool, "id", "copy-keys-from", namespace).Run()
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}

	// Set a k-v pair.
	err = exec.Command(chaosTool, "set", namespace, key, value).Run()
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}

	// Get the binary form of the namespace we created.
	listCmd := exec.Command(chaosTool, "id", "list")
	listStdout, err := listCmd.StdoutPipe()
	if err != nil {
		t.Errorf("Error creating list command: %s", err)
	}
	err = listCmd.Start()
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

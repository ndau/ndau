package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndaunode/pkg/ndau/system_vars"
	tc "github.com/oneiro-ndev/ndautool/pkg/tool/config"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

func generateMocks(ndauhome, configPath string) {
	mockPath := config.DefaultMockPath(ndauhome)
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Mock:   %s\n", mockPath)
	_, associated, err := config.MakeMock(configPath, mockPath)
	check(err)

	keys, isKeys := associated[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)
	if !isKeys {
		check(errors.New("associated data has wrong type for RFE keys"))
	}

	conf, err := tc.Load()
	check(err)
	conf.RFEKeys = keys
	err = conf.Save()
	check(err)

	os.Exit(0)
}

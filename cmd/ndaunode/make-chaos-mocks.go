package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndaunode/pkg/ndau/system_vars"
	tc "github.com/oneiro-ndev/ndautool/pkg/tool/config"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

func generateChaosMocks(ndauhome, configPath string) {
	conf, err := config.LoadConfig(configPath)
	check(err)

	associated, err := config.MakeChaosMock(conf)
	check(err)
	err = conf.Dump(configPath)
	check(err)

	keys, isKeys := associated[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)
	if !isKeys {
		check(errors.New("associated data has wrong type for RFE keys"))
	}

	tconf, err := tc.Load()
	check(err)
	tconf.RFEKeys = keys
	err = tconf.Save()
	check(err)

	os.Exit(0)
}

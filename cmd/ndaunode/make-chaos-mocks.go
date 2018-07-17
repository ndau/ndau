package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

func generateChaosMocks(ndauhome, configPath string) {
	conf, err := config.LoadDefault(configPath)
	check(err)

	associated, err := config.MakeChaosMock(conf)
	check(err)
	err = conf.Dump(configPath)
	check(err)

	keys, isKeys := associated[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)
	if !isKeys {
		check(errors.New("associated data has wrong type for RFE keys"))
	}

	generateToolConf(conf, keys)

	os.Exit(0)
}

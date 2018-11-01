package main

import (
	"fmt"
	"os"
	"time"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	tc "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	abci "github.com/tendermint/tendermint/abci/types"
)

// generate mocks, dump to a file, update tool config
func generateMocks(ndauhome, configPath string) {
	mockPath := config.DefaultMockPath(ndauhome)
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Mock:   %s\n", mockPath)
	conf, _, err := config.MakeMock(configPath, mockPath)
	check(err)

	generateToolConf(conf)

	os.Exit(0)
}

// generate mocks, dump to the chaos chain, update tool config
func generateChaosMocks(ndauhome, configPath string) {
	conf, err := config.LoadDefault(configPath)
	check(err)

	_, err = config.MakeChaosMock(conf)
	check(err)
	err = conf.Dump(configPath)
	check(err)

	generateToolConf(conf)

	os.Exit(0)
}

// generateToolConf makes necessary mock accounts and loads the tool config
//
// this function knows about which accounts need to be created, and where
// in the config their data gets saved
func generateToolConf(conf *config.Config) {
	tconf, err := tc.LoadDefault(tc.GetConfigPath())
	check(err)

	// we want to address the noms home path directly
	*useNh = true
	app, err := ndau.NewApp(getDbSpec(), "", -1, *conf)
	check(err)

	// we want to fetch the system variables, which means running
	// beginning a block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now(),
	}})
	// without the beginblock tx, we'd never actually load the mock
	// system variables into the cache

	// put RFE private keys into tool conf
	rfeAddr := address.Address{}
	err = app.System(sv.ReleaseFromEndowmentAddressName, &rfeAddr)
	check(err)
	rfePriv, err := ndau.MockSystemAccount(app, rfeAddr)
	check(err)
	tconf.RFE = &tc.SysAccount{
		Address: rfeAddr,
		Keys:    rfePriv,
	}

	// put NNR private keys into tool conf
	nnrAddr := address.Address{}
	err = app.System(sv.NominateNodeRewardAddressName, &nnrAddr)
	check(err)
	nnrPriv, err := ndau.MockSystemAccount(app, nnrAddr)
	check(err)
	tconf.NNR = &tc.SysAccount{
		Address: nnrAddr,
		Keys:    nnrPriv,
	}

	// put CVC keys into tool conf
	cvcAddr := address.Address{}
	err = app.System(sv.CommandValidatorChangeAddressName, &cvcAddr)
	check(err)
	cvcPriv, err := ndau.MockSystemAccount(app, cvcAddr)
	check(err)
	tconf.CVC = &tc.SysAccount{
		Address: cvcAddr,
		Keys:    cvcPriv,
	}

	check(tconf.Save())
}

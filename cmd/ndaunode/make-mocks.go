package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"

	metast "github.com/oneiro-ndev/metanode/pkg/meta/state"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndau/config"
	sv "github.com/oneiro-ndev/ndau/pkg/ndau/system_vars"
	tc "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	abci "github.com/tendermint/tendermint/abci/types"
)

func generateMockRFEAccounts(conf *config.Config) []address.Address {
	// we want to address the noms home path directly
	*useNh = true
	app, err := ndau.NewApp(getDbSpec(), *conf)
	check(err)
	// we want to fetch the system variables, which means running
	// beginning a block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{
		Time: time.Now().Unix(),
	}})
	// without the beginblock tx, we'd never actually load the mock
	// system variables into the cache
	rfeTransferKeys := make(sv.ReleaseFromEndowmentKeys, 0)
	err = app.System(sv.ReleaseFromEndowmentKeysName, &rfeTransferKeys)
	check(err)

	addrs := make([]address.Address, 0, len(rfeTransferKeys))

	err = app.UpdateStateImmediately(func(stateI metast.State) (metast.State, error) {
		state := stateI.(*backing.State)

		for _, trKey := range rfeTransferKeys {
			public, _, err := signature.Generate(signature.Ed25519, nil)
			if err != nil {
				return state, err
			}
			addr, err := address.Generate(address.KindEndowment, public.Bytes())
			if err != nil {
				return state, err
			}
			addrs = append(addrs, addr)

			ts, err := math.TimestampFrom(time.Now())
			if err != nil {
				return state, err
			}
			acct, _ := state.GetAccount(addr, ts)
			trKeyCopy := trKey // copy so pointers work right
			acct.TransferKeys = []signature.PublicKey{trKeyCopy}

			state.Accounts[addr.String()] = acct
		}

		return state, err
	})
	check(err)
	return addrs
}

func generateToolConf(conf *config.Config, keys []signature.PrivateKey) {
	addrs := generateMockRFEAccounts(conf)
	if len(keys) != len(addrs) {
		check(errors.New("keys and addresses have different length"))
	}

	rfeAuth := make([]tc.RFEAuth, 0, len(keys))
	for idx := 0; idx < len(keys); idx++ {
		rfeAuth = append(rfeAuth, tc.RFEAuth{
			Address: addrs[idx],
			Key:     keys[idx],
		})
	}

	tconf, err := tc.LoadDefault(tc.GetConfigPath())
	check(err)
	tconf.RFE = rfeAuth
	err = tconf.Save()
	check(err)
}

func generateMocks(ndauhome, configPath string) {
	mockPath := config.DefaultMockPath(ndauhome)
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Mock:   %s\n", mockPath)
	conf, associated, err := config.MakeMock(configPath, mockPath)
	check(err)

	keys, isKeys := associated[sv.ReleaseFromEndowmentKeysName].([]signature.PrivateKey)
	if !isKeys {
		check(errors.New("associated data has wrong type for RFE keys"))
	}

	generateToolConf(conf, keys)

	os.Exit(0)
}

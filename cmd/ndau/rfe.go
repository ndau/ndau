package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndautool/pkg/tool"
	"github.com/pkg/errors"
)

func getRfe(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s [RFE_KEY_INDEX]",
			getNdauSpec(),
			getAddressSpec(""),
		)

		getNdau := getNdauClosure(cmd)
		getAddress := getAddressClosure(cmd, "")
		index := cmd.IntArg("RFE_KEY_INDEX", 0, "index of RFE key to use to sign this transaction")

		cmd.Action = func() {
			ndauQty := getNdau()
			address := getAddress()

			if *verbose {
				fmt.Printf("Release from endowment: %s ndau to %s\n", ndauQty, address)
			}

			conf := getConfig()
			if len(conf.RFEKeys) <= *index {
				orQuit(errors.New("not enough RFE keys in configuration"))
			}
			key := conf.RFEKeys[*index]

			rfe, err := ndau.NewReleaseFromEndowment(ndauQty, address, key)
			orQuit(errors.Wrap(err, "generating Release from Endowment tx"))

			result, err := tool.ReleaseFromEndowmentCommit(tmnode(conf.Node), rfe)
			finish(*verbose, result, err, "rfe")
		}
	}
}

package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/signature/pkg/signature"
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
			if len(conf.RFE) <= *index {
				orQuit(errors.New("not enough RFE keys in configuration"))
			}

			rfe := ndau.NewReleaseFromEndowment(
				ndauQty,
				address,
				conf.RFE[*index].Address,
				sequence(conf, conf.RFE[*index].Address),
				[]signature.PrivateKey{conf.RFE[*index].Key},
			)

			result, err := tool.SendCommit(tmnode(conf.Node), &rfe)
			finish(*verbose, result, err, "rfe")
		}
	}
}

package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

func getAccountComputeEAI(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account whose delegates' EAI should be calculated")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			if *verbose {
				fmt.Printf(
					"Calculating EAI for delegates to node %s\n",
					acct.Address.String(),
				)
			}

			tx := ndau.NewComputeEAI(
				acct.Address,
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Transfer.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "compute-eai")
		}
	}
}

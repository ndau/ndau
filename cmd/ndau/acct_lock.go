package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

func getLock(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME DURATION"

		var name = cmd.StringArg("NAME", "", "Name of account to lock")
		var durationS = cmd.StringArg("DURATION", "", "Duration of notice period")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			duration, err := math.ParseDuration(*durationS)
			orQuit(err)

			if *verbose {
				fmt.Printf(
					"Locking acct %s for %s\n",
					acct.Address.String(),
					duration,
				)
			}

			tx := ndau.NewLock(
				acct.Address,
				duration,
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Transfer.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "lock")
		}
	}
}

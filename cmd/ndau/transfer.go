package main

import (
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/ndautool/pkg/tool"
	"github.com/pkg/errors"
)

func getTransfer(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s\n",
					ndauQty, from, to,
				)
			}

			conf := getConfig()

			// ensure we know the private transfer key of this account
			fromAcct, hasAcct := conf.Accounts[from.String()]
			if !hasAcct {
				orQuit(fmt.Errorf("From account '%s' not found", fromAcct.Name))
			}
			if fromAcct.Transfer == nil {
				orQuit(fmt.Errorf("From acct transfer key not set"))
			}

			// query the account to get the current sequence
			ad, _, err := tool.GetAccount(tmnode(conf.Node), from)
			orQuit(errors.Wrap(err, "Failed to get current sequence number"))

			// construct the transfer
			transfer, err := ndau.NewTransfer(from, to, ndauQty, ad.Sequence+1, fromAcct.Transfer.Private)
			orQuit(errors.Wrap(err, "Failed to construct transfer"))

			tresp, err := tool.TransferCommit(tmnode(conf.Node), *transfer)
			finish(*verbose, tresp, err, "transfer")
		}
	}
}

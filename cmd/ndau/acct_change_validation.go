package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

func getAccountChangeValidation(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account to change")

		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			public, private, err := signature.Generate(signature.Ed25519, nil)
			orQuit(errors.Wrap(err, "Failed to generate new transfer key"))
			ctk := ndau.NewChangeValidation(
				acct.Address,
				[]signature.PublicKey{public},
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Ownership.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &ctk)

			// only persist this change if there was no error
			if err == nil {
				acct.Transfer = &config.Keypair{Public: public, Private: private}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "change-transfer-key")
		}
	}
}

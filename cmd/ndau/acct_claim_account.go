package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/metanode/pkg/meta/app/code"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
	rpc "github.com/tendermint/tendermint/rpc/core/types"
)

func getAccountClaim(verbose *bool) func(*cli.Cmd) {
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

			ca := ndau.NewClaimAccount(
				acct.Address,
				acct.Ownership.Public,
				[]signature.PublicKey{public},
				acct.ValidationScript,
				sequence(conf, acct.Address),
				acct.Ownership.Private,
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &ca)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Transfer = []config.Keypair{config.Keypair{Public: public, Private: private}}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account claim")
		}
	}
}
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

func getAccountValidation(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name of account to change")

		cmd.Command(
			"reset",
			"generate a new transfer key which replaces all current transfer keys",
			getReset(verbose, name),
		)

		cmd.Command(
			"add",
			"add a new transfer key to this account",
			getAdd(verbose, name),
		)
	}
}

func getReset(verbose *bool, name *string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Transfer) == 0 {
				orQuit(errors.New("account is not yet claimed"))
			}

			public, private, err := signature.Generate(signature.Ed25519, nil)
			orQuit(errors.Wrap(err, "Failed to generate new transfer key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				[]signature.PublicKey{public},
				[]byte{},
				sequence(conf, acct.Address),
				acct.TransferPrivate(),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Transfer = []config.Keypair{config.Keypair{Public: public, Private: private}}
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation reset")
		}
	}
}

func getAdd(verbose *bool, name *string) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			if len(acct.Transfer) == 0 {
				orQuit(errors.New("account is not yet claimed"))
			}

			public, private, err := signature.Generate(signature.Ed25519, nil)
			orQuit(errors.Wrap(err, "Failed to generate new transfer key"))

			cv := ndau.NewChangeValidation(
				acct.Address,
				append(acct.TransferPublic(), public),
				[]byte{},
				sequence(conf, acct.Address),
				acct.TransferPrivate(),
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), &cv)

			// only persist this change if there was no error
			if err == nil && code.ReturnCode(resp.(*rpc.ResultBroadcastTxCommit).DeliverTx.Code) == code.OK {
				acct.Transfer = append(acct.Transfer, config.Keypair{Public: public, Private: private})
				conf.SetAccount(*acct)
				err = conf.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
			finish(*verbose, resp, err, "account validation add")
		}
	}
}

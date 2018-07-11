package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	config "github.com/oneiro-ndev/ndaunode/pkg/tool.config"
	"github.com/oneiro-ndev/ndautool/pkg/tool"
	"github.com/oneiro-ndev/signature/pkg/signature"
	"github.com/pkg/errors"
)

func getAccount(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command(
			"list",
			"list known accounts",
			getAccountList(verbose),
		)

		cmd.Command(
			"new",
			"create a new account",
			getAccountNew(verbose),
		)

		cmd.Command(
			"change-transfer-key",
			"change the account's transfer key",
			getAccountCTK(verbose),
		)

		cmd.Command(
			"query",
			"query the ndau chain about this account",
			getAccountQuery(verbose),
		)

		cmd.Command(
			"change-escrow-period",
			"change the escrow period for outbound transfers from this account",
			getAccountChangeEscrow(verbose),
		)

		cmd.Command(
			"delegate",
			"delegate EAI calculation to a node",
			getAccountDelegate(verbose),
		)

		cmd.Command(
			"compute-eai",
			"compute EAI for accounts which have delegated to this one",
			getAccountComputeEAI(verbose),
		)
	}
}

func getAccountList(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Action = func() {
			config := getConfig()
			config.EmitAccounts(os.Stdout)
		}
	}
}

func getAccountNew(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = "NAME"

		var name = subcmd.StringArg("NAME", "", "Name to associate with the new identity")

		subcmd.Action = func() {
			config := getConfig()
			err := config.CreateAccount(*name)
			orQuit(errors.Wrap(err, "Failed to create identity"))
			err = config.Save()
			orQuit(errors.Wrap(err, "saving config"))
		}
	}
}

func getAccountCTK(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = "NAME"

		var name = subcmd.StringArg("NAME", "", "Name of account to change")

		subcmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(errors.New("No such account"))
			}

			public, private, err := signature.Generate(signature.Ed25519, nil)
			orQuit(errors.Wrap(err, "Failed to generate new transfer key"))
			ctk := ndau.NewChangeTransferKey(
				acct.Address,
				public,
				ndau.SigningKeyOwnership,
				acct.Ownership.Public, acct.Ownership.Private,
			)

			resp, err := tool.ChangeTransferKeyCommit(tmnode(conf.Node), ctk)

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

func getAccountQuery(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = fmt.Sprintf("%s", getAddressSpec(""))
		getAddress := getAddressClosure(subcmd, "")

		subcmd.Action = func() {
			address := getAddress()
			config := getConfig()
			ad, resp, err := tool.GetAccount(tmnode(config.Node), address)
			if err != nil {
				finish(*verbose, resp, err, "account")
			}
			jsb, err := json.MarshalIndent(ad, "", "  ")
			fmt.Println(string(jsb))
			finish(*verbose, resp, err, "account")
		}
	}
}

func getAccountChangeEscrow(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = fmt.Sprintf(
			"NAME %s",
			getDurationSpec(),
		)

		var name = subcmd.StringArg("NAME", "", "Name of account of which to change escrow period")
		getDuration := getDurationClosure(subcmd)

		subcmd.Action = func() {
			config := getConfig()
			duration := getDuration()

			ad, hasAd := config.Accounts[*name]
			if !hasAd {
				orQuit(errors.New("No such account found"))
			}
			if ad.Transfer == nil {
				orQuit(errors.New("Address transfer key not set"))
			}

			cep, err := ndau.NewChangeEscrowPeriod(ad.Address, duration, ad.Transfer.Private)
			orQuit(errors.Wrap(err, "Creating ChangeEscrowPeriod transaction"))

			if *verbose {
				fmt.Printf(
					"Change Escrow Period for %s (%s) to %s\n",
					*name,
					ad.Address,
					duration,
				)
			}

			resp, err := tool.ChangeEscrowPeriodCommit(tmnode(config.Node), cep)
			finish(*verbose, resp, err, "change-escrow-period")
		}
	}
}

func getAccountDelegate(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = fmt.Sprintf(
			"NAME %s",
			getAddressSpec("NODE"),
		)

		var name = subcmd.StringArg("NAME", "", "Name of account whose EAI calculations should be delegated")
		getNode := getAddressClosure(subcmd, "NODE")

		subcmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			node := getNode()

			if *verbose {
				fmt.Printf(
					"Delegating %s to node %s\n",
					acct.Address.String(), node.String(),
				)
			}

			// query the account to get the current sequence
			ad, _, err := tool.GetAccount(tmnode(conf.Node), acct.Address)
			orQuit(errors.Wrap(err, "Failed to get current sequence number"))

			tx := ndau.NewDelegate(
				acct.Address, node,
				ad.Sequence+1, acct.Transfer.Private,
			)

			resp, err := tool.DelegateCommit(tmnode(conf.Node), *tx)
			finish(*verbose, resp, err, "delegate")
		}
	}
}

func getAccountComputeEAI(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = "NAME"

		var name = subcmd.StringArg("NAME", "", "Name of account whose delegates' EAI should be calculated")

		subcmd.Action = func() {
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

			// query the account to get the current sequence
			ad, _, err := tool.GetAccount(tmnode(conf.Node), acct.Address)
			orQuit(errors.Wrap(err, "Failed to get current sequence number"))

			tx := ndau.NewComputeEAI(
				acct.Address,
				ad.Sequence+1,
				acct.Transfer.Private,
			)

			resp, err := tool.ComputeEAICommit(tmnode(conf.Node), *tx)
			finish(*verbose, resp, err, "compute-eai")
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"
	math "github.com/oneiro-ndev/ndaumath/pkg/types"
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
			"change-settlement-period",
			"change the settlement period for outbound transfers from this account",
			getAccountChangeSettlement(verbose),
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

		cmd.Command(
			"lock",
			"lock this account with a specified notice period",
			getLock(verbose),
		)

		cmd.Command(
			"notify",
			"notify that this account should be unlocked once its notice period expires",
			getNotify(verbose),
		)

		cmd.Command(
			"set-rewards-target",
			"set the rewards target for this account",
			getSetRewardsTarget(verbose),
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
			ctk := ndau.NewChangeTransferKeys(
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

func getAccountChangeSettlement(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = fmt.Sprintf(
			"NAME %s",
			getDurationSpec(),
		)

		var name = subcmd.StringArg("NAME", "", "Name of account of which to change settlement period")
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

			cep, err := ndau.NewChangeSettlementPeriod(
				ad.Address,
				duration,
				sequence(config, ad.Address),
				[]signature.PrivateKey{ad.Transfer.Private},
			)
			orQuit(errors.Wrap(err, "Creating ChangeEscrowPeriod transaction"))

			if *verbose {
				fmt.Printf(
					"Change Escrow Period for %s (%s) to %s\n",
					*name,
					ad.Address,
					duration,
				)
			}

			resp, err := tool.SendCommit(tmnode(config.Node), &cep)
			finish(*verbose, resp, err, "change-settlement-period")
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

			tx := ndau.NewDelegate(
				acct.Address, node,
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Transfer.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
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

func getLock(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = "NAME DURATION"

		var name = subcmd.StringArg("NAME", "", "Name of account to lock")
		var durationS = subcmd.StringArg("DURATION", "", "Duration of notice period")

		subcmd.Action = func() {
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

func getNotify(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = "NAME"

		var name = subcmd.StringArg("NAME", "", "Name of account to lock")

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
					"Notifying acct %s\n",
					acct.Address,
				)
			}

			tx := ndau.NewNotify(
				acct.Address,
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Transfer.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "notify")
		}
	}
}

func getSetRewardsTarget(verbose *bool) func(*cli.Cmd) {
	return func(subcmd *cli.Cmd) {
		subcmd.Spec = fmt.Sprintf("NAME %s", getAddressSpec("DESTINATION"))
		getAddress := getAddressClosure(subcmd, "DESTINATION")
		var name = subcmd.StringArg("NAME", "", "Name of account to lock")

		subcmd.Action = func() {
			conf := getConfig()
			acct, hasAcct := conf.Accounts[*name]
			if !hasAcct {
				orQuit(fmt.Errorf("No such account: %s", *name))
			}
			if acct.Transfer == nil {
				orQuit(fmt.Errorf("Transfer key for %s not set", *name))
			}

			dest := getAddress()

			if *verbose {
				fmt.Printf(
					"Setting rewards target for acct %s to %s\n",
					acct.Address,
					dest,
				)
			}

			tx := ndau.NewSetRewardsTarget(
				acct.Address,
				dest,
				sequence(conf, acct.Address),
				[]signature.PrivateKey{acct.Transfer.Private},
			)

			resp, err := tool.SendCommit(tmnode(conf.Node), tx)
			finish(*verbose, resp, err, "set-rewards-target")
		}
	}
}

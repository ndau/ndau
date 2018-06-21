package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
)

func getAddressSpec() string {
	return "(NAME | -a=<ADDRESS> | --address=<ADDRESS>)"
}

func getAddressClosure(cmd *cli.Cmd) func() address.Address {
	var (
		name = cmd.StringArg("NAME", "", "Name of account")
		addr = cmd.StringOpt("a address", "", "Address")
	)

	return func() address.Address {
		if addr != nil {
			a, err := address.Validate(*addr)
			orQuit(err)
			return a
		}
		if name != nil {
			config := getConfig()
			acct, hasAcct := config.Accounts[*name]
			if hasAcct {
				return acct.Address
			}
			orQuit(errors.New("No such named account"))
		}
		orQuit(errors.New("Neither name nor address specified"))
		return address.Address{}
	}
}

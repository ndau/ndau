package main

import (
	cli "github.com/jawher/mow.cli"
	"github.com/pkg/errors"
)

func getAccountNew(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "NAME"

		var name = cmd.StringArg("NAME", "", "Name to associate with the new identity")

		cmd.Action = func() {
			config := getConfig()
			err := config.CreateAccount(*name)
			orQuit(errors.Wrap(err, "Failed to create identity"))
			err = config.Save()
			orQuit(errors.Wrap(err, "saving config"))
		}
	}
}

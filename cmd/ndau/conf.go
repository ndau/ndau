package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	config "github.com/oneiro-ndev/ndaunode/pkg/tool.config"
	"github.com/pkg/errors"
)

func getConf(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR]"

		var addr = cmd.StringArg("ADDR", config.DefaultAddress, "Address of node to connect to")

		cmd.Action = func() {
			conf, err := config.Load(config.GetConfigPath())
			if err != nil && os.IsNotExist(err) {
				conf = config.NewConfig(*addr)
			} else {
				conf.Node = *addr
			}
			err = conf.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}
	}
}

func confPath(cmd *cli.Cmd) {
	cmd.Action = func() {
		fmt.Println(config.GetConfigPath())
	}
}

package main

import (
	cli "github.com/jawher/mow.cli"
	v "github.com/oneiro-ndev/ndau/pkg/version"
)

func version(cmd *cli.Cmd) {
	cmd.Action = func() {
		v.Emit()
	}
}

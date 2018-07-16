package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("ndau", "interact with the ndau chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the ndau chain if set")
	)

	app.Command("conf", "perform initial configuration", getConf(verbose))

	app.Command("conf-path", "show location of config file", confPath)

	app.Command("account", "manage accounts", getAccount(verbose))

	app.Command("transfer", "transfer ndau from one account to another", getTransfer(verbose))

	app.Command("rfe", "release ndau from the endowment", getRfe(verbose))

	app.Command("info", "get information about node's current status", info)

	app.Command("gtvc", "send a globally trusted validator change", getGtvc(verbose))

	app.Command("server", "create server for API endpoint calls", server)

	app.Run(os.Args)
}

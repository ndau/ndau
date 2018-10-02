package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("ndau", "interact with the ndau chain")

	app.Spec = "[-v|-k]..."

	var (
		verbose = app.BoolOpt("v verbose", false, "emit detailed results from the ndau chain if set")
		keys    = app.IntOpt("k keys", 0, "bitset of keys to use in signature")
	)

	app.Command("conf", "perform initial configuration", getConf(verbose))

	app.Command("conf-path", "show location of config file", confPath)

	app.Command("account", "manage accounts", getAccount(verbose, keys))

	app.Command("transfer", "transfer ndau from one account to another", getTransfer(verbose, keys))

	app.Command("transfer-lock", "transfer ndau from one account to a new account and lock the destination", getTransferAndLock(verbose, keys))

	app.Command("rfe", "release ndau from the endowment", getRfe(verbose, keys))

	app.Command("nnr", "nominate node reward", getNNR(verbose, keys))

	app.Command("info", "get information about node's current status", info)

	app.Command("gtvc", "send a globally trusted validator change", getGtvc(verbose))

	app.Command("server", "create server for API endpoint calls", server)

	app.Run(os.Args)
}

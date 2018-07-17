package main

import (
	"encoding/hex"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getGtvc(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "PUBKEY POWER"

		pk := cmd.StringArg("PUBKEY", "", "hexadecimal encoding of ed25519 public key")
		power := cmd.IntArg("POWER", 0, "power to assign to this node")

		cmd.Action = func() {
			pkb, err := hex.DecodeString(*pk)
			orQuit(err)
			// we'd like to validate the public key length here, but we don't
			// actually know how long it should be

			if *power < 0 {
				orQuit(errors.New("negative powers not allowed in gtvc"))
			}
			power := int64(*power)

			config := getConfig()

			resp, err := tool.GTVC(tmnode(config.Node), pkb, power)
			finish(*verbose, resp, err, "gtvc")
		}
	}
}

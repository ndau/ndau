package main

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/oneiro-ndev/ndau/pkg/ndau"
	config "github.com/oneiro-ndev/ndau/pkg/tool.config"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
	"github.com/pkg/errors"
)

func getCVC(verbose *bool, keys *int) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = "(PUBKEY | -x=<PUBKEY_HEX>) POWER"

		pk64 := cmd.StringArg("PUBKEY", "", "padding-free base64 encoding of Tendermint ed25519 public key")
		pkx := cmd.StringOpt("x hex", "", "hexadecimal encoding of Tendermint ed25519 public key")
		power := cmd.IntArg("POWER", 0, "power to assign to this node")

		cmd.Action = func() {
			var pkb []byte
			var err error

			switch {
			case pkx != nil:
				pkb, err = hex.DecodeString(*pkx)
			case pk64 != nil:
				pkb, err = base64.RawStdEncoding.DecodeString(*pk64)
			default:
				err = errors.New("PUBKEY must be set")
			}
			orQuit(err)

			// we'd like to validate the public key length here, but we don't
			// actually know how long it should be

			if *power < 0 {
				orQuit(errors.New("cvc POWER must be > 0"))
			}

			conf := getConfig()
			if conf.CVC == nil {
				orQuit(errors.New("CVC data not set in tool config"))
			}

			fkeys := config.FilterK(conf.CVC.Keys, keys)

			cvc := ndau.NewCommandValidatorChange(
				pkb, int64(*power),
				sequence(conf, conf.CVC.Address),
				fkeys,
			)

			result, err := tool.SendCommit(tmnode(conf.Node), &cvc)
			finish(*verbose, result, err, "cvc")
		}
	}
}

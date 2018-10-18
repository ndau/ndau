package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getInfo(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		key := cmd.BoolOpt("k key", false, "when set, emit only the public key of the connected node")
		emithex := cmd.BoolOpt("x hex", false, "when set, emit the key as hex instead of base64")

		cmd.Spec = "[-k [-x]]"

		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node))

			if *key {
				b := info.ValidatorInfo.PubKey.Bytes()
				var p string
				if *emithex {
					p = hex.EncodeToString(b)
				} else {
					p = base64.RawStdEncoding.EncodeToString(b)
				}
				fmt.Println(p)
			} else {
				*verbose = true
			}
			finish(*verbose, info, err, "info")
		}
	}
}

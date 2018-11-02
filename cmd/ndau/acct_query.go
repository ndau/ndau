package main

import (
	"encoding/json"
	"fmt"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndau/pkg/tool"
)

func getAccountQuery(verbose *bool) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getAddressSpec(""))
		getAddress := getAddressClosure(cmd, "")

		cmd.Action = func() {
			address := getAddress()
			config := getConfig()
			ad, resp, err := tool.GetAccount(tmnode(config.Node), address)
			if err != nil {
				finish(*verbose, resp, err, "account")
			}
			if ad.LastWAAUpdate == 0 {
				// this was a marker that the account was not on the blockchain. Here, we don't care about that information
				// so restore it to its former glory
				ad.LastWAAUpdate = ad.LastEAIUpdate
			}
			jsb, err := json.MarshalIndent(ad, "", "  ")
			fmt.Println(string(jsb))
			finish(*verbose, resp, err, "account")
		}
	}
}

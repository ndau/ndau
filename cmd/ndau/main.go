package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/oneiro-ndev/ndaunode/pkg/ndau"
	"github.com/oneiro-ndev/signature/pkg/signature"

	cli "github.com/jawher/mow.cli"
	"github.com/kentquirk/boneful"
	"github.com/pkg/errors"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/oneiro-ndev/ndautool/pkg/tool"
)

func main() {
	app := cli.App("ndau", "interact with the ndau chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the ndau chain if set")
	)

	app.Command("conf", "perform initial configuration", func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR]"

		var addr = cmd.StringArg("ADDR", tool.DefaultAddress, "Address of node to connect to")

		cmd.Action = func() {
			config, err := tool.Load()
			if err != nil && os.IsNotExist(err) {
				config = tool.NewConfig(*addr)
			} else {
				config.Node = *addr
			}
			err = config.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}
	})

	app.Command("conf-path", "show location of config file", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(tool.GetConfigPath())
		}
	})

	app.Command("account", "manage accounts", func(cmd *cli.Cmd) {
		cmd.Command("list", "list known accounts", func(subcmd *cli.Cmd) {
			subcmd.Action = func() {
				config := getConfig()
				config.EmitAccounts(os.Stdout)
			}
		})

		cmd.Command("new", "create a new account", func(subcmd *cli.Cmd) {
			subcmd.Spec = "NAME"

			var name = subcmd.StringArg("NAME", "", "Name to associate with the new identity")

			subcmd.Action = func() {
				config := getConfig()
				err := config.CreateAccount(*name)
				orQuit(errors.Wrap(err, "Failed to create identity"))
				err = config.Save()
				orQuit(errors.Wrap(err, "saving config"))
			}
		})

		cmd.Command("change-transfer-key", "change the account's transfer key", func(subcmd *cli.Cmd) {
			subcmd.Spec = "NAME"

			var name = subcmd.StringArg("NAME", "", "Name of account to change")

			subcmd.Action = func() {
				config := getConfig()
				acct, hasAcct := config.Accounts[*name]
				if !hasAcct {
					orQuit(errors.New("No such account"))
				}

				public, private, err := signature.Generate(signature.Ed25519, nil)
				orQuit(errors.Wrap(err, "Failed to generate new transfer key"))
				ctk := ndau.NewChangeTransferKey(
					acct.Address,
					public,
					ndau.SigningKeyOwnership,
					acct.Ownership.Public, acct.Ownership.Private,
				)

				resp, err := tool.ChangeTransferKeyCommit(tmnode(config.Node), ctk)

				// only persist this change if there was no error
				if err == nil {
					acct.Transfer = &tool.Keypair{Public: public, Private: private}
					config.SetAccount(*acct)
					err = config.Save()
					orQuit(errors.Wrap(err, "saving config"))
				}
				finish(*verbose, resp, err, "change-transfer-key")
			}
		})
	})

	app.Command("info", "get information about node's current status", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			config := getConfig()
			info, err := tool.Info(tmnode(config.Node))
			// the whole point of this command is to get this information;
			// it makes no sense to require the verbose flag in this case
			finish(true, info, err, "info")
		}
	})

	app.Command("gtvc", "send a globally trusted validator change", func(cmd *cli.Cmd) {
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
	})

	app.Command("query-account", "query the ndau chain about this account", func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf("%s", getAddressSpec())
		getAddress := getAddressClosure(cmd)

		cmd.Action = func() {
			address := getAddress()
			config := getConfig()
			ad, resp, err := tool.GetAccount(tmnode(config.Node), address)
			if err != nil {
				finish(*verbose, resp, err, "account")
			}
			jsb, err := json.MarshalIndent(ad, "", "  ")
			fmt.Println(string(jsb))
			finish(*verbose, resp, err, "account")
		}
	})

	app.Command("server", "create server for API endpoint calls", func(cmd *cli.Cmd) {
		cmd.Spec = "PORT"

		port := cmd.StringArg("PORT", "", "port number for server to listen on")

		cmd.Action = func() {
			svc := new(boneful.Service).
				Path("/").
				Doc(`This service provides the API for Tendermint and Chaos/Order/ndau blockchain data`)

			svc.Route(svc.GET("/status").To(getStatus).
				Doc("Returns the status of the current node.").
				Operation("Status").
				Produces("application/json").
				Writes(rpctypes.ResultStatus{}))

			svc.Route(svc.GET("/health").To(getHealth).
				Doc("Returns the health of the current node.").
				Operation("Health").
				Produces("application/json").
				Writes(rpctypes.ResultHealth{}))

			svc.Route(svc.GET("/net_info").To(getNetInfo).
				Doc("Returns the network information of the current node.").
				Operation("Net Info").
				Produces("application/json").
				Writes(rpctypes.ResultNetInfo{}))

			svc.Route(svc.GET("/genesis").To(getGenesis).
				Doc("Returns the genesis block of the current node.").
				Operation("Genesis").
				Produces("application/json").
				Writes(rpctypes.ResultGenesis{}))

			svc.Route(svc.GET("/abci_info").To(getABCIInfo).
				Doc("Returns info on the ABCI interface.").
				Operation("ABCI Info").
				Produces("application/json").
				Writes(rpctypes.ResultABCIInfo{}))

			svc.Route(svc.GET("/num_unconfirmed_txs").To(getNumUnconfirmedTxs).
				Doc("Returns the number of unconfirmed transactions on the chain.").
				Operation("Num Unconfirmed Transactions").
				Produces("application/json").
				Writes(rpctypes.ResultStatus{}))

			svc.Route(svc.GET("/dump_consensus_state").To(getDumpConsensusState).
				Doc("Returns the current Tendermint consensus state in JSON").
				Operation("Dump Consensus State").
				Produces("application/json").
				Writes(rpctypes.ResultDumpConsensusState{}))

			svc.Route(svc.GET("/block").To(getBlock).
				Doc("Returns the block in the chain at the given height.").
				Operation("Get Block").
				Param(boneful.QueryParameter("height", "Height of the block in chain to return.").DataType("string").Required(true)).
				Produces("application/json").
				Writes(rpctypes.ResultBlock{}))

			svc.Route(svc.GET("/blockchain").To(getBlockChain).
				Doc("Returns a sequence of blocks starting at min_height and ending at max_height").
				Operation("Get Block Chain").
				Param(boneful.QueryParameter("min_height", "Height at which to begin retrieval of blockchain sequence.").DataType("string").Required(true)).
				Param(boneful.QueryParameter("max_height", "Height at which to end retrieval of blockchain sequence.").DataType("string").Required(true)).
				Produces("application/json").
				Writes(rpctypes.ResultBlockchainInfo{}))

			log.Printf("Chaos server listening on port %s\n", *port)
			server := &http.Server{Addr: ":" + *port, Handler: svc.Mux()}
			log.Fatal(server.ListenAndServe())
		}
	})

	app.Run(os.Args)
}

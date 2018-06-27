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
	"github.com/oneiro-ndev/ndautool/pkg/tool/config"
)

func main() {
	app := cli.App("ndau", "interact with the ndau chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the ndau chain if set")
	)

	app.Command("conf", "perform initial configuration", func(cmd *cli.Cmd) {
		cmd.Spec = "[ADDR]"

		var addr = cmd.StringArg("ADDR", config.DefaultAddress, "Address of node to connect to")

		cmd.Action = func() {
			conf, err := config.Load()
			if err != nil && os.IsNotExist(err) {
				conf = config.NewConfig(*addr)
			} else {
				conf.Node = *addr
			}
			err = conf.Save()
			orQuit(errors.Wrap(err, "Failed to save configuration"))
		}
	})

	app.Command("conf-path", "show location of config file", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			fmt.Println(config.GetConfigPath())
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
				conf := getConfig()
				acct, hasAcct := conf.Accounts[*name]
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

				resp, err := tool.ChangeTransferKeyCommit(tmnode(conf.Node), ctk)

				// only persist this change if there was no error
				if err == nil {
					acct.Transfer = &config.Keypair{Public: public, Private: private}
					conf.SetAccount(*acct)
					err = conf.Save()
					orQuit(errors.Wrap(err, "saving config"))
				}
				finish(*verbose, resp, err, "change-transfer-key")
			}
		})

		cmd.Command("query", "query the ndau chain about this account", func(subcmd *cli.Cmd) {
			subcmd.Spec = fmt.Sprintf("%s", getAddressSpec(""))
			getAddress := getAddressClosure(subcmd, "")

			subcmd.Action = func() {
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

		cmd.Command(
			"change-escrow-period",
			"change the escrow period for outbound transfers from this account",
			func(subcmd *cli.Cmd) {
				subcmd.Spec = fmt.Sprintf(
					"NAME %s",
					getDurationSpec(),
				)

				var name = subcmd.StringArg("NAME", "", "Name of account of which to change escrow period")
				getDuration := getDurationClosure(subcmd)

				subcmd.Action = func() {
					config := getConfig()
					duration := getDuration()

					ad, hasAd := config.Accounts[*name]
					if !hasAd {
						orQuit(errors.New("No such account found"))
					}
					if ad.Transfer == nil {
						orQuit(errors.New("Address transfer key not set"))
					}

					cep, err := ndau.NewChangeEscrowPeriod(ad.Address, duration, ad.Transfer.Private)
					orQuit(errors.Wrap(err, "Creating ChangeEscrowPeriod transaction"))

					if *verbose {
						fmt.Printf(
							"Change Escrow Period for %s (%s) to %s\n",
							*name,
							ad.Address,
							duration,
						)
					}

					resp, err := tool.ChangeEscrowPeriodCommit(tmnode(config.Node), cep)
					finish(*verbose, resp, err, "change-escrow-period")
				}
			},
		)
	})

	app.Command("transfer", "transfer ndau from one account to another", func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s %s",
			getNdauSpec(),
			getAddressSpec("FROM"),
			getAddressSpec("TO"),
		)

		getNdau := getNdauClosure(cmd)
		getFrom := getAddressClosure(cmd, "FROM")
		getTo := getAddressClosure(cmd, "TO")

		cmd.Action = func() {
			ndauQty := getNdau()
			from := getFrom()
			to := getTo()

			if *verbose {
				fmt.Printf(
					"Transfer %s ndau from %s to %s\n",
					ndauQty, from, to,
				)
			}

			conf := getConfig()

			// ensure we know the private transfer key of this account
			fromAcct, hasAcct := conf.Accounts[from.String()]
			if !hasAcct {
				orQuit(fmt.Errorf("From account '%s' not found", fromAcct.String()))
			}
			if fromAcct.Transfer == nil {
				orQuit(fmt.Errorf("From acct transfer key not set"))
			}

			// query the account to get the current sequence
			ad, _, err := tool.GetAccount(tmnode(conf.Node), from)
			orQuit(errors.Wrap(err, "Failed to get current sequence number"))

			// construct the transfer
			transfer, err := ndau.NewTransfer(from, to, ndauQty, ad.Sequence+1, fromAcct.Transfer.Private)
			orQuit(errors.Wrap(err, "Failed to construct transfer"))

			tresp, err := tool.TransferCommit(tmnode(conf.Node), *transfer)
			finish(*verbose, tresp, err, "transfer")
		}
	})

	app.Command("rfe", "release ndau from the endowment", func(cmd *cli.Cmd) {
		cmd.Spec = fmt.Sprintf(
			"%s %s RFE_KEY_INDEX",
			getNdauSpec(),
			getAddressSpec(""),
		)

		getNdau := getNdauClosure(cmd)
		getAddress := getAddressClosure(cmd, "")
		index := cmd.IntArg("RFE_KEY_INDEX", 0, "index of RFE key to use to sign this transaction")

		cmd.Action = func() {
			ndauQty := getNdau()
			address := getAddress()

			if *verbose {
				fmt.Printf("Release from endowment: %s ndau to %s\n", ndauQty, address)
			}

			conf := getConfig()
			if len(conf.RFEKeys) <= *index {
				orQuit(errors.New("not enough RFE keys in configuration"))
			}
			key := conf.RFEKeys[*index]

			rfe, err := ndau.NewReleaseFromEndowment(ndauQty, address, key)
			orQuit(errors.Wrap(err, "generating Release from Endowment tx"))

			result, err := tool.ReleaseFromEndowmentCommit(tmnode(conf.Node), rfe)
			finish(*verbose, result, err, "rfe")
		}
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

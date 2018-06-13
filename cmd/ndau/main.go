package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kentquirk/boneful"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/pkg/errors"

	cli "github.com/jawher/mow.cli"
	"github.com/oneiro-ndev/ndautool/pkg/tool"
)

func main() {
	app := cli.App("chaos", "interact with the chaos chain")

	app.Spec = "[-v]"

	var (
		verbose = app.BoolOpt("v verbose", false, "Emit detailed results from the chaos chain if set")
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
			/*
				router := mux.NewRouter()
				router.HandleFunc("/status", getStatus).Methods("GET")
				router.HandleFunc("/health", getHealth).Methods("GET")
				router.HandleFunc("/net_info", getNetInfo).Methods("GET")
				router.HandleFunc("/genesis", getGenesis).Methods("GET")
				router.HandleFunc("/abci_info", getABCIInfo).Methods("GET")
				router.HandleFunc("/num_unconfirmed_txs", getNumUnconfirmedTxs).Methods("GET")
				router.HandleFunc("/dump_consensus_state", getDumpConsensusState).Methods("GET")
				router.HandleFunc("/block", getBlock).Queries("height", "{height}")
				router.HandleFunc("/blockchain", getBlockChain).Queries("min_height", "{min_height}", "max_height", "{max_height}")
				router.HandleFunc("/set_key", setKeyVal).Methods("GET")
				router.HandleFunc("/get_key", getKeyVal).Methods("GET")
				router.HandleFunc("/get_ns", getNamespaces).Methods("GET")
				router.HandleFunc("/dump_key_vals", dumpKeyVals).Methods("GET")
			*/

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

			//                      router.HandleFunc("/get_key", getKeyVal).Queries("name", "{name}", "height", "{height}", "key", "{key}", "emit", "{emit}")

			/*
				"subscribe":       rpc.NewWSRPCFunc(Subscribe, "query"),
				"unsubscribe":     rpc.NewWSRPCFunc(Unsubscribe, "query"),
				"unsubscribe_all": rpc.NewWSRPCFunc(UnsubscribeAll, ""),

				// info API
				"health":               rpc.NewRPCFunc(Health, ""),
				"status":               rpc.NewRPCFunc(Status, ""),
				"net_info":             rpc.NewRPCFunc(NetInfo, ""),
				"blockchain":           rpc.NewRPCFunc(BlockchainInfo, "minHeight,maxHeight"),
				"genesis":              rpc.NewRPCFunc(Genesis, ""),
				"block":                rpc.NewRPCFunc(Block, "height"),
				"block_results":        rpc.NewRPCFunc(BlockResults, "height"),
				"commit":               rpc.NewRPCFunc(Commit, "height"),
				"tx":                   rpc.NewRPCFunc(Tx, "hash,prove"),
				"tx_search":            rpc.NewRPCFunc(TxSearch, "query,prove"),
				"validators":           rpc.NewRPCFunc(Validators, "height"),
				"dump_consensus_state": rpc.NewRPCFunc(DumpConsensusState, ""),
				"unconfirmed_txs":      rpc.NewRPCFunc(UnconfirmedTxs, ""),
				"num_unconfirmed_txs":  rpc.NewRPCFunc(NumUnconfirmedTxs, ""),

				// broadcast API
				"broadcast_tx_commit": rpc.NewRPCFunc(BroadcastTxCommit, "tx"),
				"broadcast_tx_sync":   rpc.NewRPCFunc(BroadcastTxSync, "tx"),
				"broadcast_tx_async":  rpc.NewRPCFunc(BroadcastTxAsync, "tx"),

				// abci API
				"abci_query": rpc.NewRPCFunc(ABCIQuery, "path,data,height,prove"),
				"abci_info":  rpc.NewRPCFunc(ABCIInfo, ""),
			*/
			log.Printf("Chaos server listening on port %s\n", *port)
			server := &http.Server{Addr: ":" + *port, Handler: svc.Mux()}
			log.Fatal(server.ListenAndServe())
		}
	})

	app.Run(os.Args)
}

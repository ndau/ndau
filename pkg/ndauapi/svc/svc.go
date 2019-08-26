package svc

import (
	"net/http"
	"strings"

	"github.com/tendermint/tendermint/p2p"

	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/metanode/pkg/meta/app"
	"github.com/oneiro-ndev/ndau/pkg/ndau"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndau/pkg/query"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/eai"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// NewLogMux returns a new boneful service with all of our routes and logging middleware.
func NewLogMux(cf cfg.Cfg) http.HandlerFunc {
	// Default logger that consults LOG_FORMAT and LOG_LEVEL env vars and logs to Stderr.
	cf.Logger = app.NewLogger()

	svc := New(cf)
	logmw := LogMW(svc.Mux(), cf.Logger)
	return logmw
}

// JSON is the MIME type that we process
const JSON = "application/json"

// this function is only intended for testing so it panics on errors
func keyFromString(s string) signature.PublicKey {
	var k = signature.PublicKey{}
	err := k.UnmarshalText([]byte(s))
	if err != nil {
		panic(err)
	}
	return k
}

var dummyPublic = keyFromString("npuba8jadtbbedhhdcad42tysymzpi5ec77vpi4exabh3unu2yem8wn4wv22kvvt24kpm3ghikst")
var dummyTxHash = "L4hD20bp7w4Hi19vpn46wQ"

// var dummyPublic, dummyPrivate, _ = signature.Generate(signature.Ed25519, nil)
var dummyAddress, _ = address.Generate(address.KindUser, dummyPublic.KeyBytes())
var dummyAddress2, _ = address.Generate(address.KindUser, dummyPublic.KeyBytes())
var dummyAccount = backing.AccountData{
	Balance:            123000000,
	ValidationKeys:     []signature.PublicKey{dummyPublic},
	WeightedAverageAge: 30 * types.Day,
}
var dummyTimestamp = "2018-07-10T20:01:02Z"
var dummyBlockMeta = tmtypes.BlockMeta{}
var dummyResultBlock = rpctypes.ResultBlock{
	BlockMeta: &dummyBlockMeta,
	Block:     &tmtypes.Block{},
}

func dummyParsedTimestamp() types.Timestamp {
	x, _ := types.ParseTimestamp(dummyTimestamp)
	return x
}

var dummyLockTx = ndau.NewLock(dummyAddress, 30*types.Day, 1234)

var dummyTransactionResult = routes.TransactionData{
	BlockHeight: 1234,
	TxOffset:    3,
	TxHash:      "123abc34099f",
	TxType:      "Lock",
	TxData:      dummyLockTx,
}

var dummyTransactionList = routes.TransactionList{
	Txs:        []routes.TransactionData{dummyTransactionResult},
	NextTxHash: "123abc34099f",
}

var dummySubmitResult = routes.SubmitResult{
	TxHash: "123abc34099f",
	Msg:    "only set if additional information is available",
}
var dummyPrevalidateResult = routes.PrevalidateResult{
	FeeNapu: 100,
	SibNapu: 10,
	Err:     "Err and ErrCode are only set if an error occurred",
	ErrCode: 0,
	TxHash:  "123abc34099f",
	Msg:     "only set if additional information is available",
}

// New returns a new boneful Service with routes.
func New(cf cfg.Cfg) *boneful.Service {
	svc := new(boneful.Service).
		Path("/").
		Doc(`This service provides the ndau API, used to retrieve information about and manage the ndau blockchain and
		its Tendermint consensus engine.

		It is organized into several sections:

		* /account returns data about specific accounts
		* /block returns information about blocks on the blockchain
		* /node provides information about node operations
		* /price returns information related to the ndau monetary system
		* /state provides dynamic system state information
		* /system rqueries or sets system variables
		* /transaction queries individual transactions on the blockchain
		* /tx provides tools to build, prevalidate, and submit transactions
		* /version returns current system version information

		Each of these, in turn, has several endpoints within it.
		`)

	svc.Route(svc.GET("/account/account/:address").To(routes.HandleAccount(cf)).
		Doc("Returns current state of an account given its address.").
		Notes("Will return an empty result if the account is a valid ID but not on the blockchain.").
		Operation("AccountByID").
		Produces(JSON).
		Writes(dummyAccount))

	svc.Route(svc.POST("/account/accounts").To(routes.HandleAccounts(cf)).
		Doc("Returns current state of several accounts given a list of addresses.").
		Notes("Only returns data for accounts that actively exist on the blockchain.").
		Operation("AccountsFromList").
		Consumes(JSON).
		Reads([]string{dummyAddress.String()}).
		Produces(JSON).
		Writes(map[string]backing.AccountData{dummyAddress.String(): dummyAccount}))

	svc.Route(svc.GET("/account/history/:address").To(routes.HandleAccountHistory(cf)).
		Doc("Returns the balance history of an account given its address.").
		Notes(`The history includes the timestamp, new balance, and transaction ID of each change to the account's balance.
		The result is sorted chronologically.`).
		Operation("AccountHistory").
		Param(boneful.PathParameter("address", "The address of the account for which to return history").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "The block height after which results should start.").DataType("string").Required(false)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(routes.AccountHistoryItems{Items: []routes.AccountHistoryItem{{
			Balance:   123000000,
			Timestamp: dummyTimestamp,
			TxHash:    dummyTxHash,
		}}}))

	svc.Route(svc.GET("/account/list").To(routes.HandleAccountList(cf)).
		Doc("Returns a list of account IDs.").
		Notes(`This returns a list of every account on the blockchain, sorted
		alphabetically. A maximum of 1000 accounts can be returned in a single
		request. The results are sorted by address.`).
		Operation("AccountList").
		Param(boneful.QueryParameter("after", "The address after which (sorted alphabetically) results should start.").DataType("string").Required(false)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(query.AccountListQueryResponse{
			NumAccounts: 1,
			FirstIndex:  1,
			After:       dummyAddress.String(),
			NextAfter:   dummyAddress2.String(),
			Accounts:    []string{dummyAddress.String()},
		}))

	svc.Route(svc.GET("/account/currencyseats").To(routes.HandleAccountCurrencySeats(cf)).
		Doc("Returns a list of ndau 'currency seats', the oldest 3000 accounts containing more than 1000 ndau.").
		Notes(`The ndau currency seats are accounts containing more than 1000 ndau. The seniority of
		a currency seat is determined by how long it has been above the 1000 threshold, so this endpoint
		also sorts the result by age (oldest first). It does not return detailed account information.`).
		Operation("AccountCurrencySeats").
		Param(boneful.QueryParameter("limit", "The max number of items to return (default=3000)").DataType("int").Required(false)).
		Produces(JSON).
		Writes(query.AccountListQueryResponse{
			NumAccounts: 1,
			FirstIndex:  1,
			After:       dummyAddress.String(),
			NextAfter:   dummyAddress.String(),
			Accounts:    []string{dummyAddress.String()},
		}))

	svc.Route(svc.POST("/account/eai/rate").To(routes.GetEAIRate(cf)).
		Operation("DEPRECATED:AccountEAIRate").
		Doc("This call is deprecated -- please use /system/eai/rate.").
		Consumes(JSON).
		Produces(JSON))

	svc.Route(svc.GET("/block/before/:height").To(routes.HandleBlockBefore(cf)).
		Operation("BlockBefore").
		Doc("Returns a (possibly filtered) sequence of block metadata for blocks on or before a given height.").
		Param(boneful.PathParameter("height", "Blocks greater than this height will not be returned.").DataType("int").Required(true)).
		Param(boneful.QueryParameter("filter", "Set to 'noempty' to exclude empty blocks.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "The block height after which no more results should be returned.").DataType("int").Required(false)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(rpctypes.ResultBlockchainInfo{
			LastHeight: 12345,
			BlockMetas: []*tmtypes.BlockMeta{&dummyBlockMeta},
		}))

	svc.Route(svc.GET("/block/current").To(routes.HandleBlockHeight(cf)).
		Operation("BlockCurrent").
		Doc("Returns the most recent block in the chain").
		Produces(JSON).
		Writes(dummyResultBlock))

	svc.Route(svc.GET("/block/hash/:blockhash").To(routes.HandleBlockHash(cf)).
		Operation("BlockHash").
		Doc("Returns the block in the chain with the given hash.").
		Param(boneful.PathParameter("blockhash", "Hex hash of the block in chain to return.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(dummyResultBlock))

	svc.Route(svc.GET("/block/height/:height").To(routes.HandleBlockHeight(cf)).
		Operation("BlockHeight").
		Doc("Returns the block in the chain at the given height.").
		Param(boneful.PathParameter("height", "Height of the block in chain to return.").DataType("int").Required(true)).
		Produces(JSON).
		Writes(dummyResultBlock))

	svc.Route(svc.GET("/block/range/:first/:last").To(routes.HandleBlockRange(cf)).
		Operation("BlockRange").
		Doc("Returns a sequence of block metadata starting at first and ending at last").
		Param(boneful.PathParameter("first", "Height at which to begin retrieval of blocks.").DataType("int").Required(true)).
		Param(boneful.PathParameter("last", "Height at which to end retrieval of blocks.").DataType("int").Required(true)).
		Param(boneful.QueryParameter("noempty", "Set to nonblank value to exclude empty blocks").DataType("string").Required(true)).
		Produces(JSON).
		Writes(rpctypes.ResultBlockchainInfo{
			LastHeight: 12345,
			BlockMetas: []*tmtypes.BlockMeta{&dummyBlockMeta},
		}))

	svc.Route(svc.GET("/block/transactions/:height").To(routes.HandleBlockTransactions(cf)).
		Operation("BlockTransactions").
		Doc("Returns transaction hashes for a given block. These can be used to fetch data for individual transactions.").
		Param(boneful.PathParameter("height", "Height of the block in chain containing transactions.").DataType("int").Required(true)).
		Produces(JSON).
		Writes([]string{dummyTxHash}))

	svc.Route(svc.GET("/block/daterange/:first/:last").To(routes.HandleBlockDateRange(cf)).
		Operation("BlockDateRange").
		Doc("Returns a sequence of block metadata starting at first date and ending at last date").
		Param(boneful.PathParameter("first", "Timestamp (ISO 3339) at which to begin (inclusive) retrieval of blocks.").DataType("string").Required(true)).
		Param(boneful.PathParameter("last", "Timestamp (ISO 3339) at which to end (exclusive) retrieval of blocks.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("noempty", "Set to nonblank value to exclude empty blocks").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "The timestamp after which results should start (use the last value from the previous page).").DataType("string").Required(false)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(rpctypes.ResultBlockchainInfo{
			LastHeight: 12345,
			BlockMetas: []*tmtypes.BlockMeta{&dummyBlockMeta},
		}))

	svc.Route(svc.GET("/node/status").To(routes.GetStatus(cf)).
		Operation("NodeStatus").
		Doc("Returns the status of the current node.").
		Produces(JSON).
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/node/health").To(routes.GetHealth(cf)).
		Operation("NodeHealth").
		Doc("Returns the health of the current node by doing a simple test for connectivity and response.").
		Produces(JSON).
		Writes(routes.HealthResponse{}))

	svc.Route(svc.GET("/node/net").To(routes.GetNetInfo(cf)).
		Operation("NodeNetInfo").
		Doc("Returns the network information of the current node.").
		Produces(JSON).
		Writes(rpctypes.ResultNetInfo{}))

	svc.Route(svc.GET("/node/genesis").To(routes.GetGenesis(cf)).
		Operation("NodeGenesis").
		Doc("Returns the genesis document of the current node.").
		Produces(JSON).
		Writes(rpctypes.ResultGenesis{}))

	svc.Route(svc.GET("/node/abci").To(routes.GetABCIInfo(cf)).
		Operation("NodeABCIInfo").
		Doc("Returns info on the node's ABCI interface.").
		Produces(JSON).
		Writes(rpctypes.ResultABCIInfo{}))

	svc.Route(svc.GET("/node/consensus").To(routes.GetDumpConsensusState(cf)).
		Operation("NodeConsensusState").
		Doc("Returns the current Tendermint consensus state in JSON").
		Produces(JSON).
		Writes(rpctypes.ResultDumpConsensusState{}))

	svc.Route(svc.GET("/node/nodes").To(routes.GetNodeList(cf)).
		Operation("NodeList").
		Doc("Returns a list of all nodes.").
		Produces(JSON).
		Writes(routes.ResultNodeList{}))

	svc.Route(svc.GET("/node/:id").To(routes.GetNode(cf)).
		Operation("NodeID").
		Doc("Returns a single node.").
		Param(boneful.PathParameter("id", "the NodeID as a hex string")).
		Produces(JSON).
		Writes(p2p.NodeInfo.NetAddress))

	svc.Route(svc.GET("/price/current").To(routes.GetPriceData(cf)).
		Operation("PriceInfo").
		Doc("Returns current price data for key parameters.").
		Notes(`Returns current price information:
		* Market price
		* Target price
		* Total ndau issued from the endowment
		* Total ndau in circulation
		* Total SIB burned
		* Current SIB in effect
		`).
		Produces(JSON).
		Writes(routes.PriceInfo{
			MarketPrice: 1234 * 1000000000,
			TargetPrice: 5678 * 1000000000,
			TotalIssued: 2919000 * 100000000,
			TotalNdau:   3141593 * 100000000,
			TotalSIB:    123 * 100000000,
			CurrentSIB:  9876543210,
		}))

	svc.Route(svc.GET("/price/height/:height").To(routes.HandlePriceHeight(cf)).
		Operation("PriceHeight").
		Doc("Returns the collection of price data as of a specific ndau block height.").
		Param(boneful.PathParameter("height", "Height from the ndau chain.").DataType("int").Required(true)).
		Produces(JSON).
		Writes(routes.PriceInfo{}))

	svc.Route(svc.GET("/price/history").To(routes.HandlePriceHistory(cf)).
		Operation("PriceHistory").
		Doc("Returns an array of data from the order chain at periodic intervals over time, sorted chronologically.").
		Param(boneful.QueryParameter("limit", "Maximum number of values to return; default=100, max=1000.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("period", "Duration between samples (ex: 1d, 5m); default=1d.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("before", "Timestamp (ISO 8601) to end (exclusive); default=now.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "Timestamp (ISO 8601) to start (inclusive); default=before-(limit*period).").DataType("string").Required(true)).
		Produces(JSON).
		Writes([]routes.PriceHistoryRecord{}))

	svc.Route(svc.GET("/state/delegates").To(routes.HandleStateDelegates(cf)).
		Operation("StateDelegates").
		Doc("Returns the current collection of delegate information.").
		Produces(JSON).
		Writes(""))

	svc.Route(svc.GET("/system/all").To(routes.HandleSystemAll(cf)).
		Operation("SystemAll").
		Doc("Returns the names and current values of all currently-defined system variables.").
		Produces(JSON).
		Writes(""))

	svc.Route(svc.GET("/system/get/:sysvars").To(routes.HandleSystemGet(cf)).
		Doc("Return the names and current values of some currently definted system variables.").
		Operation("SystemGet").
		Param(boneful.PathParameter("sysvars", "A comma-separated list of system variables of interest.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(""))

	svc.Route(svc.POST("/system/set/:sysvar").To(routes.HandleSystemSet(cf)).
		Doc("Returns a transaction which sets a system variable.").
		Notes(`The body of the request accepts JSON and heuristically transforms
		it into the data format used internally on the blockchain. Do not use any sort
		of wrapping object. The correct structure of the object to send depends on
		the system variable in question.

		Returns the JSON encoding of a SetSysvar transaction. It is the caller's
		responsibility to update this transaction with appropriate sequence and
		signatures and then send it at the normal endpoint (/tx/submit/setsysvar).`).
		Operation("SystemSet").
		Param(boneful.PathParameter("sysvar", "The name of the system variable to return").DataType("string").Required(true)).
		Consumes(JSON).
		Produces(JSON).
		Writes(""))

	svc.Route(svc.GET("/system/history/:sysvar").To(routes.HandleSystemHistory(cf)).
		Doc("Returns the value history of a system variable given its name.").
		Notes(`The history includes the height and value of each change to the system variable.
		The result is sorted chronologically.`).
		Operation("SystemHistory").
		Param(boneful.PathParameter("sysvar", "The name of the system variable for which to return history").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "The block height after which results should start.").DataType("string").Required(false)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(query.SysvarHistoryResponse{History: []query.SysvarHistoricalValue{{
			Height: 12345,
			Value:  []byte("Value"),
		}}}))

	svc.Route(svc.POST("/system/eai/rate").To(routes.GetEAIRate(cf)).
		Operation("SystemEAIRate").
		Doc("Returns eai rates for a collection of account information.").
		Notes(`Accepts an array of rate requests that includes an address
		field; this field may be any string (the account information is not
		checked). It returns an array of rate responses, which includes
		the address passed so that responses may be correctly correlated
		to the input.

		It accepts a timestamp, which will be used to adjust WAA in the
		event the account is locked and has a non-nil "unlocksOn" value.
		If the timestamp field is omitted, the current time is used.

		EAIRate in the response is an integer equal to the fractional EAI
		rate times 10^12.
		`).
		Consumes(JSON).
		Reads([]routes.EAIRateRequest{routes.EAIRateRequest{
			Address: dummyAddress.String(),
			WAA:     90 * types.Day,
			Lock:    *backing.NewLock(180*types.Day, eai.DefaultLockBonusEAI),
			At:      dummyParsedTimestamp(),
		}}).
		Produces(JSON).
		Writes([]routes.EAIRateResponse{routes.EAIRateResponse{
			Address: dummyAddress.String(),
			EAIRate: 60000000000,
		}}))

	svc.Route(svc.GET("/transaction/:txhash").To(routes.HandleTransactionFetch(cf)).
		Doc("Returns a transaction from the blockchain given its tx hash.").
		Operation("TransactionByHash").
		Produces(JSON).
		Writes(dummyTransactionResult))

	svc.Route(svc.GET("/transaction/before/:txhash").To(routes.HandleTransactionBefore(cf)).
		Operation("TransactionBefore").
		Doc("Returns a sequence of transaction metadata for transactions equal to or before a given transaction.").
		Param(boneful.PathParameter("txhash", "Transactions after this will not be returned.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("limit", "The maximum number of items to return. Use a positive limit, or 0 for getting max results; default=0, max=100").DataType("int").Required(false)).
		Produces(JSON).
		Writes(dummyTransactionList))

	svc.Route(svc.POST("/tx/prevalidate/:txtype").To(routes.HandlePrevalidateTx(cf)).
		Doc("Prevalidates a transaction (tells if it would be accepted and what the transaction fee will be.").
		Notes("Transactions consist of JSON for any defined transaction type (see submit).").
		Operation("TxPrevalidate").
		Consumes(JSON).
		Reads(dummyLockTx).
		Produces(JSON).
		Writes(dummyPrevalidateResult))

	svc.Route(svc.POST("/tx/submit/:txtype").To(routes.HandleSubmitTx(cf)).
		Doc("Submits a transaction.").
		Notes("Transactions consist of JSON for any defined transaction type. Valid transaction names and aliases are: " + strings.Join(routes.TxNames(), ", ")).
		Operation("TxSubmit").
		Consumes(JSON).
		Reads(dummyLockTx).
		Produces(JSON).
		Writes(dummySubmitResult))

	svc.Route(svc.GET("/version").To(routes.HandleVersion(cf)).
		Doc("Delivers version information").
		Operation("Version").
		Produces(JSON).
		Writes(routes.VersionResult{
			NdauVersion: "v1.2.3",
			NdauSha:     "23abc35",
			Network:     "mainnet",
		}))
	return svc
}

// Add call to get list of nodes

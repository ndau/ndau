package svc

import (
	"net/http"

	"github.com/oneiro-ndev/ndaumath/pkg/address"

	"github.com/tendermint/tendermint/p2p"

	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	"github.com/oneiro-ndev/signature/pkg/signature"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// NewLogMux returns a new boneful service with all of our routes and logging middleware.
func NewLogMux(cf cfg.Cfg) http.HandlerFunc {
	svc := New(cf)
	return LogMW(svc.Mux())
}

// JSON is the MIME type that we process
const JSON = "application/json"

var dummyPublic, dummyPrivate, _ = signature.Generate(signature.Ed25519, nil)
var dummyAddress, _ = address.Generate(address.KindUser, dummyPublic.Bytes())
var dummyAccount = backing.AccountData{
	Balance:            123000000,
	TransferKeys:       []signature.PublicKey{dummyPublic},
	WeightedAverageAge: 30 * types.Day,
}
var dummyTimestamp = "2018-07-18T20:01:02Z"
var dummyBlockMeta = tmtypes.BlockMeta{}
var dummyResultBlock = rpctypes.ResultBlock{
	BlockMeta: &dummyBlockMeta,
	Block:     &tmtypes.Block{},
}
var dummyPreparedTx = routes.PreparedTx{
	TxData:        "base64 tx data",
	SignableBytes: "base64 bytes to be signed",
	Signature:     "base64 signature of SignableBytes",
}
var dummyTxResult = routes.TxResult{
	TxHash: "123abc34099f",
}

// New returns a new boneful Service with routes.
func New(cf cfg.Cfg) *boneful.Service {
	svc := new(boneful.Service).
		Path("/").
		Doc(`This service provides the API for Tendermint and Chaos/Order/ndau blockchain data.

		It is organized into several sections:

		* /account returns data about specific accounts
		* /block returns information about blocks on the blockchain
		* /chaos returns information from the chaos chain
		* /node provides information about node operations
		* /order returns information from the order chain
		* /transaction allows querying individual transactions on the blockchain
		* /tx provides tools to build and submit transactions

		Each of these, in turn, has several endpoints within it.
		`)

	svc.Route(svc.GET("/account/account/:address").To(routes.HandleAccount(cf)).
		Doc("Returns current state of an account given its address.").
		Operation("AccountByID").
		Produces(JSON).
		Writes(routes.AccountDataResponse{
			AccountData: dummyAccount,
			Address:     dummyAddress.String(),
		}))

	svc.Route(svc.POST("/account/accounts").To(routes.HandleAccounts(cf)).
		Doc("Returns current state of several accounts given a list of addresses.").
		Operation("AccountsFromList").
		Consumes(JSON).
		Reads([]string{dummyAddress.String()}).
		Produces(JSON).
		Writes(routes.AccountResponse{
			AcctData: []routes.AccountDataResponse{routes.AccountDataResponse{
				AccountData: dummyAccount,
				Address:     dummyAddress.String(),
			}}}))

	svc.Route(svc.POST("/account/eai/rate").To(routes.GetEAIRate(cf)).
		Operation("AccountEAIRate").
		Doc("Returns eai rates for a collection of account information.").
		Notes(`Accepts an array of rate requests that includes an address
		field; this field may be any string (the account information is not
		checked). It returns an array of rate responses, which includes
		the address passed so that responses may be correctly correlated
		to the input.
		`).
		Consumes(JSON).
		Reads([]routes.EAIRateRequest{routes.EAIRateRequest{
			Address: dummyAddress.String(),
			WAA:     90 * types.Day,
			Lock:    backing.Lock{NoticePeriod: 180 * types.Day},
		}}).
		Produces(JSON).
		Writes([]routes.EAIRateResponse{routes.EAIRateResponse{
			Address: dummyAddress.String(),
			EAIRate: 6000000,
		}}))

	svc.Route(svc.GET("/account/history/:accountid").To(routes.HandleAccount(cf)).
		Doc("Returns the balance history of an account given its address.").
		Notes(`The history includes the timestamp, new balance, and transaction ID of each change to the account's balance.
		The result is reverse sorted chronologically from the current time, and supports paging by time.`).
		Operation("AccountByID").
		Param(boneful.QueryParameter("limit", "Maximum number of transactions to return; default=10.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("before", "Timestamp (ISO 8601) to start looking backwards; default=now.").DataType("string").Required(true)).
		Produces(JSON).
		Writes([]routes.BalanceHistoryItem{
			routes.BalanceHistoryItem{
				Balance:   123000000,
				Timestamp: dummyTimestamp,
				TxHash:    "abc123def456",
			},
		}))

	svc.Route(svc.GET("/block/hash/:blockhash").To(routes.HandleBlockHash(cf)).
		Operation("BlockHash").
		Doc("Returns the block in the chain with the given hash.").
		Param(boneful.QueryParameter("blockhash", "Hash of the block in chain to return.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(dummyResultBlock))

	svc.Route(svc.GET("/block/height/:height").To(routes.HandleBlockHeight(cf)).
		Operation("BlockHeight").
		Doc("Returns the block in the chain at the given height.").
		Param(boneful.QueryParameter("height", "Height of the block in chain to return.").DataType("int").Required(true)).
		Produces(JSON).
		Writes(dummyResultBlock))

	svc.Route(svc.GET("/block/range/:first/:last").To(routes.HandleBlockRange(cf)).
		Operation("BlockRange").
		Doc("Returns a sequence of blocks starting at first and ending at last").
		Param(boneful.PathParameter("first", "Height at which to begin retrieval of blocks.").DataType("int").Required(true)).
		Param(boneful.PathParameter("last", "Height at which to end retrieval of blocks.").DataType("int").Required(true)).
		Produces(JSON).
		Writes(rpctypes.ResultBlockchainInfo{
			BlockMetas: []*tmtypes.BlockMeta{&dummyBlockMeta},
		}))

	svc.Route(svc.GET("/chaos/system/names").To(routes.HandleBlockRange(cf)).
		Operation("ChaosSystemNames").
		Doc("Returns all current named system variables on the chaos chain.").
		Produces(JSON).
		Writes(""))

	svc.Route(svc.GET("/chaos/system/:key").To(routes.HandleBlockRange(cf)).
		Operation("ChaosSystemKey").
		Doc("Returns the current value of a system variable from the chaos chain.").
		Param(boneful.PathParameter("key", "Name of the system variable.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(""))

	svc.Route(svc.GET("/chaos/history/:key").To(routes.HandleBlockRange(cf)).
		Operation("ChaosHistoryKey").
		Doc("Returns the history of changes to a value of a chaos chain system variable.").
		Notes(`The history includes the timestamp, new value, and transaction ID of each change to the account's balance.
		The result is reverse sorted chronologically from the current time, and supports paging by time.`).
		Param(boneful.PathParameter("key", "Name of the system variable.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("limit", "Maximum number of values to return; default=10.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("before", "Timestamp (ISO 8601) to start looking backwards; default=now.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(routes.ChaosHistoryResponse{}))

	svc.Route(svc.GET("/node/status").To(routes.GetStatus(cf)).
		Operation("NodeStatus").
		Doc("Returns the status of the current node.").
		Produces(JSON).
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/node/health").To(routes.GetHealth(cf)).
		Operation("NodeHealth").
		Doc("Returns the health of the current node.").
		Produces(JSON).
		Writes(rpctypes.ResultHealth{}))

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

	svc.Route(svc.GET("/node/unconfirmed").To(routes.HandleNumUnconfirmedTxs(cf)).
		Operation("NodeNumUnconfirmedTransactions").
		Doc("Returns the number of unconfirmed transactions on the chain.").
		Produces(JSON).
		Writes(rpctypes.ResultStatus{}))

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
		Writes(p2p.NodeInfo{}))

	svc.Route(svc.GET("/order/hash/:ndauhash").To(routes.HandleOrderHash(cf)).
		Operation("OrderHash").
		Doc("Returns the collection of data from the order chain as of a specific ndau blockhash.").
		Param(boneful.PathParameter("ndauhash", "Hash from the ndau chain.").DataType("string").Required(true)).
		Produces(JSON).
		Writes(routes.OrderChainInfo{}))

	svc.Route(svc.GET("/order/height/:ndauheight").To(routes.HandleOrderHeight(cf)).
		Operation("OrderHeight").
		Doc("Returns the collection of data from the order chain as of a specific ndau block height.").
		Param(boneful.PathParameter("ndauheight", "Height from the ndau chain.").DataType("int").Required(true)).
		Produces(JSON).
		Writes(routes.OrderChainInfo{}))

	svc.Route(svc.GET("/order/history/").To(routes.HandleOrderHistory(cf)).
		Operation("OrderHistory").
		Doc("Returns an array of data from the order chain at periodic intervals over time, sorted chronologically.").
		Param(boneful.QueryParameter("limit", "Maximum number of values to return; default=100, max=1000.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("period", "Duration between samples (ex: 1d, 5m); default=1d.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("before", "Timestamp (ISO 8601) to end (exclusive); default=now.").DataType("string").Required(true)).
		Param(boneful.QueryParameter("after", "Timestamp (ISO 8601) to start (inclusive); default=before-(limit*period).").DataType("string").Required(true)).
		Produces(JSON).
		Writes([]routes.OrderHistoryRecord{}))

	svc.Route(svc.GET("/order/current").To(routes.GetOrderChainData(cf)).
		Operation("OrderCurrent").
		Doc("Returns current order chain data for key parameters.").
		Notes(`Returns current order chain information for 5 parameters:
		* Market price
		* Target price
		* Floor price
		* Total ndau sold from the endowment
		* Total ndau in circulation
		`).
		Produces(JSON).
		Writes(routes.OrderChainInfo{
			MarketPrice:   16.85,
			TargetPrice:   17.00,
			FloorPrice:    2.57,
			EndowmentSold: 2919000 * 100000000,
			TotalNdau:     3141593 * 100000000,
			PriceUnits:    "USD",
		}))

	svc.Route(svc.GET("/transaction/:txhash").To(routes.HandleTransactionFetch(cf)).
		Doc("Returns a transaction given its tx hash.").
		Operation("TransactionByHash").
		Produces(JSON).
		Writes(routes.TransactionData{}))

	svc.Route(svc.POST("/tx/changevalidation").To(routes.HandleChangeValidation(cf)).
		Doc("Returns a prepared ChangeValidation transaction for signature.").
		Operation("TxChangeValidation").
		Consumes(JSON).
		Reads(routes.TxChangeValidationRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/changesettlement").To(routes.HandleChangeSettlement(cf)).
		Doc("Returns a prepared ChangeSettlement transaction for signature.").
		Operation("TxChangeSettlement").
		Consumes(JSON).
		Reads(routes.TxChangeSettlementRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/claimaccount").To(routes.HandleClaimAccount(cf)).
		Doc("Returns a prepared ClaimAccount transaction for signature.").
		Operation("TxClaimAccount").
		Consumes(JSON).
		Reads(routes.TxClaimAccountRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/claimnoderewards").To(routes.HandleClaimNodeRewards(cf)).
		Doc("Returns a prepared ClaimNodeRewards transaction for signature.").
		Operation("TxClaimNodeRewards").
		Consumes(JSON).
		Reads(routes.TxClaimNodeRewardsRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/crediteai").To(routes.HandleCreditEAI(cf)).
		Doc("Returns a prepared CreditEAI transaction for signature.").
		Operation("TxCreditEAI").
		Consumes(JSON).
		Reads(routes.TxCreditEAIRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/delegate").To(routes.HandleDelegate(cf)).
		Doc("Returns a prepared Delegate transaction for signature.").
		Operation("TxDelegate").
		Consumes(JSON).
		Reads(routes.TxDelegateRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/lock").To(routes.HandleLock(cf)).
		Doc("Returns a prepared Lock transaction for signature.").
		Operation("TxLock").
		Consumes(JSON).
		Reads(routes.TxLockRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/nominatenodereward").To(routes.HandleNominateNodeReward(cf)).
		Doc("Returns a prepared NominateNodeReward transaction for signature.").
		Operation("TxNominateNodeReward").
		Consumes(JSON).
		Reads(routes.TxNominateNodeRewardRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/notify").To(routes.HandleNotify(cf)).
		Doc("Returns a prepared Notify transaction for signature.").
		Operation("TxNotify").
		Consumes(JSON).
		Reads(routes.TxNotifyRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/registernode").To(routes.HandleRegisterNode(cf)).
		Doc("Returns a prepared RegisterNode transaction for signature.").
		Operation("TxRegisterNode").
		Consumes(JSON).
		Reads(routes.TxRegisterNodeRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/releasefromendowment").To(routes.HandleReleaseFromEndowment(cf)).
		Doc("Returns a prepared ReleaseFromEndowment transaction for signature.").
		Operation("TxReleaseFromEndowment").
		Consumes(JSON).
		Reads(routes.TxReleaseFromEndowmentRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/setrewardsdest").To(routes.HandleSetRewardsDest(cf)).
		Doc("Returns a prepared SetRewardsDest transaction for signature.").
		Operation("TxSetRewardsDest").
		Consumes(JSON).
		Reads(routes.TxSetRewardsDestRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/stake").To(routes.HandleStake(cf)).
		Doc("Returns a prepared Stake transaction for signature.").
		Operation("TxStake").
		Consumes(JSON).
		Reads(routes.TxStakeRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/transfer").To(routes.HandleTransfer(cf)).
		Doc("Returns a prepared Transfer transaction for signature.").
		Operation("TxTransfer").
		Consumes(JSON).
		Reads(routes.TxTransferRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/transferandlock").To(routes.HandleTransferAndLock(cf)).
		Doc("Returns a prepared TransferAndLock	transaction for signature.").
		Operation("TxTransferAndLock").
		Consumes(JSON).
		Reads(routes.TxTransferAndLockRequest{}).
		Produces(JSON).
		Writes(routes.PreparedTx{}))

	svc.Route(svc.POST("/tx/submit").To(routes.HandleSubmitTx(cf)).
		Doc("Submits a prepared transaction.").
		Operation("TxSubmit").
		Consumes(JSON).
		Reads(dummyPreparedTx).
		Produces(JSON).
		Writes(dummyTxResult))
	return svc
}

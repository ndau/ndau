package svc

import (
	"net/http"

	"github.com/tendermint/tendermint/p2p"

	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/ndau/pkg/ndau/backing"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/oneiro-ndev/ndaumath/pkg/types"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// NewLogMux returns a new boneful service with all of our routes and logging middleware.
func NewLogMux(cf cfg.Cfg) http.HandlerFunc {
	svc := New(cf)
	return LogMW(svc.Mux())
}

// New returns a new boneful Service with routes.
func New(cf cfg.Cfg) *boneful.Service {
	svc := new(boneful.Service).
		Path("/").
		Doc(`This service provides the API for Tendermint and Chaos/Order/ndau blockchain data`)

	svc.Route(svc.GET("/status").To(routes.GetStatus(cf)).
		Operation("Status").
		Doc("Returns the status of the current node.").
		Produces("application/json").
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/health").To(routes.GetHealth(cf)).
		Operation("Health").
		Doc("Returns the health of the current node.").
		Produces("application/json").
		Writes(rpctypes.ResultHealth{}))

	svc.Route(svc.GET("/net").To(routes.GetNetInfo(cf)).
		Operation("NetInfo").
		Doc("Returns the network information of the current node.").
		Produces("application/json").
		Writes(rpctypes.ResultNetInfo{}))

	svc.Route(svc.GET("/genesis").To(routes.GetGenesis(cf)).
		Operation("Genesis").
		Doc("Returns the genesis block of the current node.").
		Produces("application/json").
		Writes(rpctypes.ResultGenesis{}))

	svc.Route(svc.GET("/abci").To(routes.GetABCIInfo(cf)).
		Operation("ABCIInfo").
		Doc("Returns info on the ABCI interface.").
		Produces("application/json").
		Writes(rpctypes.ResultABCIInfo{}))

	svc.Route(svc.GET("/unconfirmed").To(routes.GetNumUnconfirmedTxs(cf)).
		Operation("NumUnconfirmedTransactions").
		Doc("Returns the number of unconfirmed transactions on the chain.").
		Produces("application/json").
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/consensus").To(routes.GetDumpConsensusState(cf)).
		Operation("DumpConsensusState").
		Doc("Returns the current Tendermint consensus state in JSON").
		Produces("application/json").
		Writes(rpctypes.ResultDumpConsensusState{}))

	svc.Route(svc.GET("/block").To(routes.GetBlock(cf)).
		Operation("GetBlock").
		Doc("Returns the block in the chain at the given height.").
		Param(boneful.QueryParameter("height", "Height of the block in chain to return.").DataType("string").Required(true)).
		Produces("application/json").
		Writes(rpctypes.ResultBlock{}))

	svc.Route(svc.GET("/blockchain").To(routes.GetBlockchain(cf)).
		Operation("GetBlockChain").
		Doc("Returns a sequence of blocks starting at min_height and ending at max_height").
		Param(boneful.QueryParameter(routes.StartKey, "Height at which to begin retrieval of blockchain sequence.").DataType("string").Required(true)).
		Param(boneful.QueryParameter(routes.EndKey, "Height at which to end retrieval of blockchain sequence.").DataType("string").Required(true)).
		Produces("application/json").
		Writes(rpctypes.ResultBlockchainInfo{}))

	svc.Route(svc.GET("/nodes").To(routes.GetNodeList(cf)).
		Operation("NodeList").
		Doc("Returns a list of all nodes.").
		Produces("application/json").
		Writes(routes.ResultNodeList{}))

	svc.Route(svc.GET("/nodes/:id").To(routes.GetNode(cf)).
		Operation("NodeID").
		Doc("Returns a single node.").
		Param(boneful.PathParameter("id", "the NodeID as a hex string")).
		Produces("application/json").
		Writes(p2p.NodeInfo{}))

	svc.Route(svc.POST("/accounts").To(routes.GetAccount(cf)).
		Doc("Returns a list of addresses.").
		Operation("Address List").
		Produces("application/json").
		Writes(routes.AccountResponse{}))

	svc.Route(svc.POST("/eai/rate").To(routes.GetEAIRate(cf)).
		Operation("EAIRate").
		Doc("Returns eai rates for a collection of account information.").
		Notes(`Accepts an array of rate requests that includes an address
		field; this field may be any string (the account information is not
		checked). It returns an array of rate responses, which includes
		the address passed so that responses may be correctly correlated
		to the input.
		`).
		Consumes("application/json").
		Reads([]routes.EAIRateRequest{routes.EAIRateRequest{
			Address: "accountAddress",
			WAA:     90 * types.Day,
			Lock:    backing.Lock{NoticePeriod: 180 * types.Day},
		}}).
		Produces("application/json").
		Writes([]routes.EAIRateResponse{routes.EAIRateResponse{
			Address: "accountAddress",
			EAIRate: 6000000,
		}}))

	svc.Route(svc.GET("/order/current").To(routes.GetOrderChainData(cf)).
		Operation("CurrentOrderData").
		Doc("Returns current order chain data for key parameters.").
		Notes(`Returns current order chain information for 5 parameters:
		* Market price
		* Target price
		* Floor price
		* Total ndau sold from the endowment
		* Total ndau in circulation
		`).
		Produces("application/json").
		Writes(routes.OrderChainInfo{
			MarketPrice:   16.85,
			TargetPrice:   17.00,
			FloorPrice:    2.57,
			EndowmentSold: 2919000 * 100000000,
			TotalNdau:     3141593 * 100000000,
			PriceUnits:    "USD",
		}))

	return svc
}

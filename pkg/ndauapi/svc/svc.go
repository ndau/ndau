package svc

import (
	"net/http"

	"github.com/tendermint/tendermint/p2p"

	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
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
		Doc("Returns the status of the current node.").
		Operation("Status").
		Produces("application/json").
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/health").To(routes.GetHealth(cf)).
		Doc("Returns the health of the current node.").
		Operation("Health").
		Produces("application/json").
		Writes(rpctypes.ResultHealth{}))

	svc.Route(svc.GET("/net").To(routes.GetNetInfo(cf)).
		Doc("Returns the network information of the current node.").
		Operation("Net Info").
		Produces("application/json").
		Writes(rpctypes.ResultNetInfo{}))

	svc.Route(svc.GET("/genesis").To(routes.GetGenesis(cf)).
		Doc("Returns the genesis block of the current node.").
		Operation("Genesis").
		Produces("application/json").
		Writes(rpctypes.ResultGenesis{}))

	svc.Route(svc.GET("/abci").To(routes.GetABCIInfo(cf)).
		Doc("Returns info on the ABCI interface.").
		Operation("ABCI Info").
		Produces("application/json").
		Writes(rpctypes.ResultABCIInfo{}))

	svc.Route(svc.GET("/unconfirmed").To(routes.GetNumUnconfirmedTxs(cf)).
		Doc("Returns the number of unconfirmed transactions on the chain.").
		Operation("Num Unconfirmed Transactions").
		Produces("application/json").
		Writes(rpctypes.ResultStatus{}))

	svc.Route(svc.GET("/consensus").To(routes.GetDumpConsensusState(cf)).
		Doc("Returns the current Tendermint consensus state in JSON").
		Operation("Dump Consensus State").
		Produces("application/json").
		Writes(rpctypes.ResultDumpConsensusState{}))

	svc.Route(svc.GET("/block").To(routes.GetBlock(cf)).
		Doc("Returns the block in the chain at the given height.").
		Operation("Get Block").
		Param(boneful.QueryParameter("height", "Height of the block in chain to return.").DataType("string").Required(true)).
		Produces("application/json").
		Writes(rpctypes.ResultBlock{}))

	svc.Route(svc.GET("/blockchain").To(routes.GetBlockchain(cf)).
		Doc("Returns a sequence of blocks starting at min_height and ending at max_height").
		Operation("Get Block Chain").
		Param(boneful.QueryParameter(routes.StartKey, "Height at which to begin retrieval of blockchain sequence.").DataType("string").Required(true)).
		Param(boneful.QueryParameter(routes.EndKey, "Height at which to end retrieval of blockchain sequence.").DataType("string").Required(true)).
		Produces("application/json").
		Writes(rpctypes.ResultBlockchainInfo{}))

	svc.Route(svc.GET("/nodes").To(routes.GetNodeList(cf)).
		Doc("Returns a list of all nodes.").
		Operation("Node List").
		Produces("application/json").
		Writes(routes.ResultNodeList{}))

	svc.Route(svc.GET("/nodes/:id").To(routes.GetNode(cf)).
		Doc("Returns a single node.").
		Operation("Node List").
		Produces("application/json").
		Writes(p2p.NodeInfo{}))
	return svc
}

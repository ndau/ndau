package sdk

import (
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/p2p"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Info gets the node's current status
func (c *Client) Info() (status *rpctypes.ResultStatus, err error) {
	err = c.get(status, c.URL("node/status"))
	err = errors.Wrap(err, "fetching node status from API")
	return
}

// Info gets the node's current status
func Info(node *Client) (status *rpctypes.ResultStatus, err error) {
	return node.Info()
}

// Health is a simple check of a node's health
func (c *Client) Health() (resp *routes.HealthResponse, err error) {
	err = c.get(resp, c.URL("node/health"))
	err = errors.Wrap(err, "fetching node health from API")
	return
}

// NetInfo returns the network information of the node
func (c *Client) NetInfo() (ni *rpctypes.ResultNetInfo, err error) {
	err = c.get(ni, c.URL("node/net"))
	err = errors.Wrap(err, "fetching node net info from API")
	return
}

// Genesis returns the genesis document of the node
func (c *Client) Genesis() (g *rpctypes.ResultGenesis, err error) {
	err = c.get(g, c.URL("node/genesis"))
	err = errors.Wrap(err, "fetching node genesis document from API")
	return
}

// ABCIInfo returns the node's ABCI data
func (c *Client) ABCIInfo() (abci *rpctypes.ResultABCIInfo, err error) {
	err = c.get(abci, c.URL("node/abci"))
	err = errors.Wrap(err, "fetching node ABCI info from API")
	return
}

// Consensus return's the node's current Tendermint consensus state
func (c *Client) Consensus() (cs *rpctypes.ResultConsensusState, err error) {
	err = c.get(cs, c.URL("node/consensus"))
	err = errors.Wrap(err, "fetching node consensus state from API")
	return
}

// Nodes returns a list of all nodes
func (c *Client) Nodes() (nodes *routes.ResultNodeList, err error) {
	err = c.get(nodes, c.URL("node/nodes"))
	err = errors.Wrap(err, "fetching node list from API")
	return
}

// Node returns a single node
func (c *Client) Node(id string) (node *p2p.DefaultNodeInfo, err error) {
	err = c.get(node, c.URL("node/%s", id))
	err = errors.Wrap(err, "fetching node from API")
	return
}

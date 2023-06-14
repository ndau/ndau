package sdk

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"github.com/ndau/ndau/pkg/ndauapi/routes"
	"github.com/pkg/errors"
	rpctypes "github.com/oneiro-ndev/tendermint.0.32.3/rpc/core/types"
)

// Info gets the node's current status
func (c *Client) Info() (status *rpctypes.ResultStatus, err error) {
	status = new(rpctypes.ResultStatus)
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
	resp = new(routes.HealthResponse)
	err = c.get(resp, c.URL("node/health"))
	err = errors.Wrap(err, "fetching node health from API")
	return
}

// NetInfo returns the network information of the node
func (c *Client) NetInfo() (ni *rpctypes.ResultNetInfo, err error) {
	ni = new(rpctypes.ResultNetInfo)
	err = c.get(ni, c.URL("node/net"))
	err = errors.Wrap(err, "fetching node net info from API")
	return
}

// Genesis returns the genesis document of the node
func (c *Client) Genesis() (g *rpctypes.ResultGenesis, err error) {
	g = new(rpctypes.ResultGenesis)
	err = c.get(g, c.URL("node/genesis"))
	err = errors.Wrap(err, "fetching node genesis document from API")
	return
}

// ABCIInfo returns the node's ABCI data
func (c *Client) ABCIInfo() (abci *rpctypes.ResultABCIInfo, err error) {
	abci = new(rpctypes.ResultABCIInfo)
	err = c.get(abci, c.URL("node/abci"))
	err = errors.Wrap(err, "fetching node ABCI info from API")
	return
}

// Consensus return's the node's current Tendermint consensus state
func (c *Client) Consensus() (cs *rpctypes.ResultConsensusState, err error) {
	cs = new(rpctypes.ResultConsensusState)
	err = c.get(cs, c.URL("node/consensus"))
	err = errors.Wrap(err, "fetching node consensus state from API")
	return
}

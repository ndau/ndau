package mock

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"testing"

	"github.com/ndau/ndau/pkg/ndau"
	"github.com/ndau/ndau/pkg/ndau/config"
	"github.com/ndau/ndau/pkg/ndauapi/cfg"
	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmmock "github.com/tendermint/tendermint/rpc/client/mock"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Client returns a TMClient connected to a mock tendermint connected to a
// real but empty ndau.App connected to an in-memory noms.
func Client(t *testing.T, fixtures ...func(abcitypes.Application)) cfg.TMClient {
	ndauconf, err := config.DefaultConfig()
	require.NoError(t, err)
	app, err := ndau.NewAppSilent("mem", "", -1, *ndauconf)
	require.NoError(t, err)

	c := client{
		tmmock.ABCIApp{
			App: app,
		},
	}

	for _, fixture := range fixtures {
		fixture(c.App)
	}

	return c
}

type client struct {
	tmmock.ABCIApp
}

// Block implements TMClient
func (client) Block(height *int64) (*rpctypes.ResultBlock, error) {
	return &rpctypes.ResultBlock{}, nil
}

// BlockchainInfo implements TMClient
func (client) BlockchainInfo(int64, int64) (*rpctypes.ResultBlockchainInfo, error) {
	return &rpctypes.ResultBlockchainInfo{}, nil
}

// ConsensusState implements TMClient
func (client) ConsensusState() (*rpctypes.ResultConsensusState, error) {
	return &rpctypes.ResultConsensusState{}, nil
}

// DumpConsensusState implements TMClient
func (client) DumpConsensusState() (*rpctypes.ResultDumpConsensusState, error) {
	return &rpctypes.ResultDumpConsensusState{}, nil
}

// Genesis implements TMClient
func (client) Genesis() (*rpctypes.ResultGenesis, error) {
	return &rpctypes.ResultGenesis{}, nil
}

// Health implements TMClient
func (client) Health() (*rpctypes.ResultHealth, error) {
	return &rpctypes.ResultHealth{}, nil
}

// NetInfo implements TMClient
func (client) NetInfo() (*rpctypes.ResultNetInfo, error) {
	return &rpctypes.ResultNetInfo{}, nil
}

// Status implements TMClient
func (client) Status() (*rpctypes.ResultStatus, error) {
	return &rpctypes.ResultStatus{}, nil
}

var _ cfg.TMClient = (*client)(nil)

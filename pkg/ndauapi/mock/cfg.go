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

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/writers/pkg/testwriter"
	"github.com/sirupsen/logrus"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

// Cfg creates a mock config
//
// This configuration is connected to a mock tendermint client, which in turn
// is connected to a real but empty ndau App, which uses an in-memory database
func Cfg(t *testing.T, fixtures ...func(abcitypes.Application)) cfg.Cfg {
	l := logrus.New()
	l.SetOutput(testwriter.New(t))

	return cfg.Cfg{
		Node:   Client(t, fixtures...),
		Port:   3030,
		Logger: l,
	}
}

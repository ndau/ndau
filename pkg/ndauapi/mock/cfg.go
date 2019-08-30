package mock

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

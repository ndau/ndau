package cfg

import (
	"fmt"
	"os"
	"strconv"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/ws"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/rpc/client"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	defaultPort = 3030
)

// A TMClient implements all the tendermint client methods we expect
//
// It should be relatively easy to implement by wrapping
// https://godoc.org/github.com/tendermint/tendermint/rpc/client/mock#ABCIApp
// with something which mocks a block
type TMClient interface {
	client.ABCIClient
	client.HistoryClient
	client.NetworkClient
	client.StatusClient
	Block(height *int64) (*rpctypes.ResultBlock, error)
}

// Cfg represents this application's configuration.
type Cfg struct {
	Node   TMClient // connects to TM
	Port   int      // Port is which port to listen for connections.
	Logger logrus.FieldLogger
}

// NewFromEnv initializes configuration and returns a config struct
func NewFromEnv() (Cfg, []string, error) {
	url := os.Getenv("NDAUAPI_NDAU_RPC_URL")
	if url == "" {
		return Cfg{}, nil, errors.New("NDAUAPI_NDAU_RPC_URL must be set")
	}
	return New(url)
}

// New initializes configuration and returns a config struct
func New(nodeAddr string) (Cfg, []string, error) {
	var warn []string

	node, err := ws.Node(nodeAddr)
	if err != nil {
		return Cfg{}, nil, errors.Wrap(err, "connecting to TM node")
	}

	cf := Cfg{
		// get configuration from env vars
		Node:   node,
		Port:   0,
		Logger: logrus.New(),
	}

	// validate
	strPort := os.Getenv("NDAUAPI_PORT")
	if strPort == "" {
		cf.Port = defaultPort
	} else {
		port, err := strconv.Atoi(strPort)
		if err != nil {
			return cf, warn, fmt.Errorf("cannot use value '%s' for port: %v", strPort, err)
		}
		cf.Port = port
	}
	if !(cf.Port > 1024 && cf.Port < 65535) {
		return cf, warn, fmt.Errorf("port (%v) must be within the user or dynamic/private range. (1024-65535)", cf.Port)
	}

	return cf, warn, nil
}

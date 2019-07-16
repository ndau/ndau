package cfg

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort = 3030
)

// Cfg represents this application's configuration.
type Cfg struct {
	NodeAddress string // NodeAddress is the Tendermint IP and RPC port of the ndau node.
	Port        int    // Port is which port to listen for connections.
}

// New initializes config and returns the singleton config struct if already initialized.
func New() (Cfg, []string, error) {

	var warn []string

	cf := Cfg{
		// get configuration from env vars
		NodeAddress: os.Getenv("NDAUAPI_NDAU_RPC_URL"),
		Port:        0,
	}

	if cf.NodeAddress == "" {
		return cf, nil, errors.New("NDAUAPI_NDAU_RPC_URL is required")
	}

	// validate
	strPort := os.Getenv("NDAUAPI_PORT")
	if strPort == "" {
		cf.Port = 3030
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

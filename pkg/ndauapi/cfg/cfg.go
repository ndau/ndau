package cfg

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	defaultLogLevel = log.InfoLevel
	defaultPort     = 3030
)

// Cfg reporesents this application's configuration.
type Cfg struct {
	LogLevel    string // LogLevel is translated into a constant value from logrus.
	NodeAddress string // NodeAddress is the Tendermint IP and RPC port of the ndau node.
	Port        int    // Port is which port to listen for connections.
}

// New initializes config and returns the singleton config struct if already initialized.
func New() (Cfg, []string, error) {

	var warn []string

	cf := Cfg{
		// get configuration from env vars
		LogLevel:    os.Getenv("NDAUAPI_LOG_LEVEL"),
		NodeAddress: os.Getenv("NDAUAPI_NODE_ADDRESS"),
		Port:        0,
	}

	// use defaults if necessary
	if cf.LogLevel == "" {
		warn = append(warn, fmt.Sprintf("Using default log level: %v", defaultLogLevel))
	}

	if cf.NodeAddress == "" {
		return cf, nil, errors.New("NDAUAPI_NODE_ADDRESS is required")
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

	// Set the log level
	switch cf.LogLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(defaultLogLevel)
	}
	cf.LogLevel = string(defaultLogLevel)

	return cf, warn, nil
}

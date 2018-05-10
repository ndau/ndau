package config

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// NamespacedKey is a namespace and key which together identify a unique value on the chaos chain.
type NamespacedKey struct {
	Namespace B64Data
	Key       B64Data
}

// Config defines configuration data for the ndau node
type Config struct {
	// ChaosAddress is the address of a chaos node.
	//
	// This is used to query the chaos chain for relevant data.
	// It is normally a fully qualified HTTP or TCP address.
	ChaosAddress string

	// UseMock is normally an empty string.
	//
	// When UseMock is not empty, it must be the path to a toml file
	// which contains the mock chain data. In this case, the mock data
	// overrides the actual chain data; the actual chain is not queried.
	UseMock string

	// SystemVariableIndirect is a namespaced key at which the master system
	// variable indirection map is located.
	//
	// This is very sensitive information! Changing it will lead to a fork!
	// Users should never change this! On the other hand, it's still better
	// to have it defined as a configuration variable than to hardcode it.
	//
	// The value on the chaos chain to which this points must be the
	// serialized Protobuf encoding of a SVIMap.
	SystemVariableIndirect NamespacedKey
}

// DefaultConfigPath returns the default path at which a config file is expected
func DefaultConfigPath(ndauhome string) string {
	return filepath.Join(ndauhome, "ndau", "config.toml")
}

// LoadConfig returns a config object loaded from its file
func LoadConfig(configPath string) (*Config, error) {
	config := new(Config)
	_, err := toml.DecodeFile(configPath, config)
	return config, err
}

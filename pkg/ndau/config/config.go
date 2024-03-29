package config

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ndau/ndaumath/pkg/address"
	"github.com/ndau/ndaumath/pkg/signature"
)

// A Keypair holds a pair of keys
type Keypair struct {
	Public  signature.PublicKey
	Private signature.PrivateKey
}

// Node holds node configuration data
type Node struct {
	Ownership Keypair
	Address   address.Address
}

// Config defines configuration data for the ndau node
type Config struct {
	// Node contains node configuration data
	Node Node

	// NodeRewardWebhook, if set, must be a URL.
	//
	// If set, then in the course of the NominateNodeReward transaction,
	// the node will create a POST request to this URL with a
	// JSON body:
	//
	// {
	//     "random": <int from Nominate tx>,
	//     "winner": <string address of winning node>
	// }
	//
	// This allows node operators to respond appropriately when their own node
	// wins, so they can create a `ClaimNodeReward` transaction.
	NodeRewardWebhook *string

	// NodeRewardWebhookDelay is a floating-point number of seconds.
	//
	// If set, the node will wait a random number of seconds between 0 and this
	// value before dispatching its request to the Node Reward Webhook. This
	// is because many nodes may all dispatch these requests to the same webhook;
	// having an interval between them helps humans understand the sequencing.
	//
	// This is a pointer for optionality: if a pointer is not present in a
	// TOML file, the unmarshaler just lets it be nil. If a non-pointer value
	// is not present, it complains about the missing value. Missing values
	// are treated as 0: no delay.
	NodeRewardWebhookDelay *float64

	// Map whose keys are features,
	// and whose values are the mainnet block height at which the feature becomes active.
	Features map[string]uint64
}

// DefaultConfig creates a new config object with sensible defaults
func DefaultConfig() (*Config, error) {
	config := new(Config)
	public, private, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		return config, err
	}
	config.Node.Ownership.Public = public
	config.Node.Ownership.Private = private
	addr, err := address.Generate(address.KindUser, public.KeyBytes())
	if err != nil {
		return config, err
	}
	config.Node.Address = addr
	return config, nil
}

// DefaultConfigPath returns the default path at which a config file is expected
func DefaultConfigPath(ndauhome string) string {
	return filepath.Join(ndauhome, "ndau", "config.toml")
}

// Load returns a config object loaded from its file
func Load(configPath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, os.ErrNotExist
	}
	config := new(Config)
	err = toml.Unmarshal(bytes, config)
	return config, err
}

// LoadDefault returns a config object loaded from its file
//
// If the file does not exist, a default is transparently created
func LoadDefault(configPath string) (*Config, error) {
	config, err := Load(configPath)

	// EJM - 2023-02-07 Failure to load the config file should be a fatal error, not a signal to enable all features
	// all the time. If a node with all (known) features enabled is desired, just create a config file and set all
	// feature heights to 0. Setting all features active because the config file is in the wrong place leads to
	// errors (app hash mismatches) that are difficult to debug.

	if err != nil {
		return nil, err
	}

	return config, err
}

// Dump writes the given config object to the specified file
func (c *Config) Dump(configPath string) error {
	// if the parent directories of this config don't exist, make them
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}
	fp, err := os.Create(configPath)
	defer fp.Close()
	if err != nil {
		return err
	}
	return toml.NewEncoder(fp).Encode(c)
}

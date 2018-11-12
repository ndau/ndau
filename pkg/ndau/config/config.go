package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/oneiro-ndev/chaos/pkg/genesisfile"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/oneiro-ndev/system_vars/pkg/svi"
	"github.com/pkg/errors"
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
	// ChaosAddress is the address of a chaos node.
	//
	// This is used to query the chaos chain for relevant data.
	// It is normally a fully qualified HTTP or TCP address.
	ChaosAddress string

	// When UseMock is set and not empty, it must be the path to a toml file
	// which contains the mock chain data. In this case, the mock data
	// overrides the actual chain data; the actual chain is not queried.
	UseMock *string

	// SystemVariableIndirect is a namespaced key at which the master system
	// variable indirection map is located.
	//
	// This is very sensitive information! Changing it will lead to a fork!
	// Users should never change this! On the other hand, it's still better
	// to have it defined as a configuration variable than to hardcode it.
	SystemVariableIndirect svi.Location

	// ChaosTimeout is the time in milliseconds which should be allowed
	// for reads from the chaos chain.
	//
	// Because all system variables must be fetched from the chaos chain
	// every block, this should be shorter than the block time for
	// the ndau chain.
	ChaosTimeout int

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
}

// DefaultConfig creates a new config object with sensible defaults
func DefaultConfig() (*Config, error) {
	config := new(Config)
	config.ChaosTimeout = 500
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
	if err != nil && os.IsNotExist(err) {
		config, err = DefaultConfig()
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

// UpdateFrom updates the config file from the given genesisfile path
func (c *Config) UpdateFrom(gfilepath string) error {
	// check all potential error conditions before modifying c
	gfile, err := genesisfile.Load(gfilepath)
	if err != nil {
		return errors.Wrap(err, "loading genesisfile")
	}
	sviLoc, err := gfile.FindSVIStub()
	if err != nil {
		return errors.Wrap(err, "searching for svi stub")
	}
	if sviLoc == nil {
		return errors.New("svi stub not present in genesisfile")
	}

	gfileabspath, err := filepath.Abs(gfilepath)
	if err != nil {
		return errors.Wrap(err, "finding genesisfile absolute path")
	}

	// update c
	c.SystemVariableIndirect = *sviLoc
	c.UseMock = &gfileabspath
	return nil
}

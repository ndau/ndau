package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/signature/pkg/signature"
)

// DefaultAddress of the node to connect to
const DefaultAddress string = "http://localhost:46657"

// GetConfigPath returns the location at which configuration is stored
func GetConfigPath() string {
	ndauhome := os.ExpandEnv("$NDAUHOME")
	if len(ndauhome) == 0 {
		home := os.ExpandEnv("$HOME")
		ndauhome = path.Join(home, ".ndau")
	}
	return path.Join(ndauhome, "ndau", "ndautool.toml")
}

// Config represents all data from `ndautool.toml`
type Config struct {
	Node string

	Accounts map[string]*Account

	// for non-node-operators, this will always be empty, but for testing
	// purposes we want the ndau tool to be able to issue release from
	// endowment transactions, so it needs to know about these.
	RFEKeys []signature.PrivateKey
}

// NewConfig creates a new configuration with the given address
func NewConfig(node string) *Config {
	return &Config{
		Node:     node,
		Accounts: make(map[string]*Account, 0),
	}
}

// DefaultConfig creates a new configuration with the default address
func DefaultConfig() *Config {
	return NewConfig(DefaultAddress)
}

// Load the current configuration
func Load() (*Config, error) {
	var tconfig tomlConfig
	_, err := toml.DecodeFile(GetConfigPath(), &tconfig)
	if err != nil {
		return nil, err
	}

	config, err := tconfig.toConfig()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Save the current configuration
func (c *Config) Save() error {
	tc, err := c.toToml()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(tc); err != nil {
		return err
	}
	cp := GetConfigPath()
	dir, _ := filepath.Split(cp)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(cp, buf.Bytes(), 0600) // u=rw;go-
}

// GetAccounts returns the unique accounts known to this config.
//
// Accounts are stored by default under two keys: one for the address,
// and one for the name. That's hard to iter over, so this excludes the
// duplicates.
func (c Config) GetAccounts() []*Account {
	accts := make([]*Account, 0, len(c.Accounts))
	for key, acct := range c.Accounts {
		if key == acct.Address.String() {
			accts = append(accts, acct)
		}
	}
	return accts
}

// EmitAccounts writes the accounts to the supplied writer
func (c Config) EmitAccounts(w io.Writer) {
	for _, acct := range c.GetAccounts() {
		fmt.Fprintln(w, acct)
	}
}

// CreateAccount creates an account with the given name
func (c *Config) CreateAccount(name string) error {
	public, private, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		return errors.Wrap(err, "generate keypair")
	}
	addr, err := address.Generate(address.KindUser, public.Bytes())
	if err != nil {
		return errors.Wrap(err, "generate address")
	}
	acct := Account{
		Name:      name,
		Address:   addr,
		Ownership: Keypair{Public: public, Private: private},
	}
	c.SetAccount(acct)
	return nil
}

// SetAccount sets the appropriate keys for the given account
func (c *Config) SetAccount(acct Account) {
	c.Accounts[acct.Name] = &acct
	c.Accounts[acct.Address.String()] = &acct
}

func (c Config) toToml() (tomlConfig, error) {
	tacs := make([]tomlAccount, 0, len(c.Accounts))
	for _, acct := range c.GetAccounts() {
		tac, err := acct.toToml()
		if err != nil {
			return tomlConfig{}, err
		}
		tacs = append(tacs, tac)

	}
	var rfes []string // default nil, so unset in toml
	if len(c.RFEKeys) > 0 {
		rfes = make([]string, 0, len(c.RFEKeys))
		for _, rfe := range c.RFEKeys {
			rfeb, err := rfe.Marshal()
			if err != nil {
				return tomlConfig{}, err
			}
			rfes = append(rfes, base64.StdEncoding.EncodeToString(rfeb))
		}
	}
	return tomlConfig{
		Node:     c.Node,
		Accounts: tacs,
		RFEKeys:  rfes,
	}, nil
}

// Config represents all data from `ndautool.toml`
type tomlConfig struct {
	Node     string        `toml:"node"`
	Accounts []tomlAccount `toml:"accounts"`
	RFEKeys  []string      `toml:"rfe_keys"`
}

func (tc tomlConfig) toConfig() (Config, error) {
	acts := make(map[string]*Account, 2*len(tc.Accounts))
	for _, tac := range tc.Accounts {
		act, err := tac.toAccount()
		if err != nil {
			return Config{}, err
		}
		acts[act.Address.String()] = &act
		if act.Name != "" {
			acts[act.Name] = &act
		}
	}
	rfes := make([]signature.PrivateKey, 0, len(tc.RFEKeys))
	for _, keyb64 := range tc.RFEKeys {
		bytes, err := base64.StdEncoding.DecodeString(keyb64)
		if err != nil {
			return Config{}, err
		}
		pk := signature.PrivateKey{}
		err = pk.Unmarshal(bytes)
		if err != nil {
			return Config{}, err
		}
		rfes = append(rfes, pk)
	}
	return Config{
		Node:     tc.Node,
		Accounts: acts,
		RFEKeys:  rfes,
	}, nil
}

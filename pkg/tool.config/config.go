package config

import (
	"bytes"
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
	Node     string              `toml:"node"`
	Accounts map[string]*Account `toml:"accounts"`

	// for non-node-operators, this will always be empty, but for testing
	// purposes we want the ndau tool to be able to issue release from
	// endowment transactions, so it needs to know about these.
	RFEKeys []signature.PrivateKey `toml:"rfe_keys"`
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

// Load returns a config object loaded from its file
func Load(configPath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, os.ErrNotExist
	}
	tconfig := new(tomlConfig)
	err = toml.Unmarshal(bytes, tconfig)
	if err != nil {
		return nil, err
	}

	return tconfig.toConfig()
}

// LoadDefault returns a config object loaded from its file
//
// If the file does not exist, a default is transparently created
func LoadDefault(configPath string) (*Config, error) {
	config, err := Load(configPath)
	if err != nil && os.IsNotExist(err) {
		config = DefaultConfig()
		err = nil
	}
	return config, err
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
	tacs := make([]Account, 0, len(c.Accounts))
	for _, acct := range c.GetAccounts() {
		tacs = append(tacs, *acct)
	}

	return tomlConfig{
		Node:     c.Node,
		Accounts: tacs,
		RFEKeys:  c.RFEKeys,
	}, nil
}

// Config represents all data from `ndautool.toml`
type tomlConfig struct {
	Node     string                 `toml:"node"`
	Accounts []Account              `toml:"accounts"`
	RFEKeys  []signature.PrivateKey `toml:"rfe_keys"`
}

func (tc tomlConfig) toConfig() (*Config, error) {
	acts := make(map[string]*Account, 2*len(tc.Accounts))
	for _, act := range tc.Accounts {
		acts[act.Address.String()] = &act
		if act.Name != "" {
			acts[act.Name] = &act
		}
	}

	return &Config{
		Node:     tc.Node,
		Accounts: acts,
		RFEKeys:  tc.RFEKeys,
	}, nil
}

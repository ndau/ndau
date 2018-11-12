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

	"github.com/oneiro-ndev/system_vars/pkg/system_vars"

	"github.com/BurntSushi/toml"
	generator "github.com/oneiro-ndev/chaos_genesis/pkg/genesis.generator"
	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/oneiro-ndev/ndaumath/pkg/signature"
	"github.com/pkg/errors"
)

// DefaultAddress of the node to connect to
const DefaultAddress string = "http://localhost:26657"

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
	RFE      *SysAccount         `toml:"rfe"`
	NNR      *SysAccount         `toml:"nnr"`
	CVC      *SysAccount         `toml:"cvc"`
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
	if _, found := c.Accounts[name]; found {
		return errors.New("account already exists: " + name)
	}
	public, private, err := signature.Generate(signature.Ed25519, nil)
	if err != nil {
		return errors.Wrap(err, "generate keypair")
	}
	addr, err := address.Generate(address.KindUser, public.KeyBytes())
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
		RFE:      c.RFE,
		NNR:      c.NNR,
		CVC:      c.CVC,
	}, nil
}

// Config represents all data from `ndautool.toml`
type tomlConfig struct {
	Node     string      `toml:"node"`
	Accounts []Account   `toml:"accounts"`
	RFE      *SysAccount `toml:"rfe"`
	NNR      *SysAccount `toml:"nnr"`
	CVC      *SysAccount `toml:"cvc"`
}

func (tc tomlConfig) toConfig() (*Config, error) {
	acts := make(map[string]*Account, 2*len(tc.Accounts))
	for _, act := range tc.Accounts {
		// &act is the same address on every iteration,
		// so just using it will cause logic errors.
		// copying to a new address should work, though.
		acct := act
		acts[act.Address.String()] = &acct
		if act.Name != "" {
			acts[act.Name] = &acct
		}
	}

	return &Config{
		Node:     tc.Node,
		Accounts: acts,
		RFE:      tc.RFE,
		NNR:      tc.NNR,
		CVC:      tc.CVC,
	}, nil
}

// UpdateFrom updates the config file given the path to the associated data file
// and the public key of the BPC.
func (c *Config) UpdateFrom(asscPath string, bpcPublic []byte) error {
	asscFile := make(generator.AssociatedFile)
	_, err := toml.DecodeFile(asscPath, &asscFile)
	if err != nil {
		return errors.Wrap(err, "reading asscfile path")
	}
	bpcs := base64.StdEncoding.EncodeToString(bpcPublic)
	assc, ok := asscFile[bpcs]
	if !ok {
		return errors.New("bpc key not found in associated data file")
	}

	sysaccts := []struct {
		configAcct **SysAccount
		sys        sv.SysAcct
	}{
		{&c.CVC, sv.CommandValidatorChange},
		{&c.NNR, sv.NominateNodeReward},
		{&c.RFE, sv.ReleaseFromEndowment},
	}
	for _, sa := range sysaccts {
		*sa.configAcct, err = SysAccountFromAssc(assc, sa.sys)
		if err != nil {
			return err
		}
	}

	return nil
}

package tool

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
}

// NewConfig creates a new configuration with the given address
func NewConfig(node string) *Config {
	return &Config{
		Node:     node,
		Accounts: make(map[string]*Account, 0),
	}
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
	return tomlConfig{
		Node:     c.Node,
		Accounts: tacs,
	}, nil
}

// Config represents all data from `ndautool.toml`
type tomlConfig struct {
	Node string

	Accounts []tomlAccount
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
	return Config{
		Node:     tc.Node,
		Accounts: acts,
	}, nil
}

// An Account contains the data necessary to interact with an account:
//
// ownership keys, transfer keys if assigned, an account nickname, and an address
type Account struct {
	Name      string
	Address   address.Address
	Ownership Keypair
	Transfer  *Keypair
}

func (acct Account) toToml() (tomlAccount, error) {
	ownership, err := acct.Ownership.toToml()
	if err != nil {
		return tomlAccount{}, errors.Wrap(err, "ownership")
	}
	var transfer *tomlKeypair
	if acct.Transfer != nil {
		tr, err := acct.Transfer.toToml()
		if err != nil {
			return tomlAccount{}, errors.Wrap(err, "transfer")
		}
		transfer = &tr
	}
	return tomlAccount{
		Name:      acct.Name,
		Address:   acct.Address.String(),
		Ownership: ownership,
		Transfer:  transfer,
	}, nil
}

// String satisfies io.Stringer
func (acct Account) String() string {
	var id string
	if acct.Name != "" {
		id = acct.Name
	} else {
		id = acct.Address.String()
	}
	return fmt.Sprintf("%s: owner %s transfer %s", id, acct.Ownership, acct.Transfer)
}

// tomlAccount is an account being prepared for toml marshaling
type tomlAccount struct {
	Name      string
	Address   string
	Ownership tomlKeypair
	Transfer  *tomlKeypair
}

func (ta tomlAccount) toAccount() (Account, error) {
	ownership, err := ta.Ownership.toKeypair()
	if err != nil {
		return Account{}, errors.Wrap(err, "ownership")
	}

	var transfer *Keypair
	if ta.Transfer != nil {
		tr, err := ta.Transfer.toKeypair()
		if err != nil {
			return Account{}, errors.Wrap(err, "transfer")
		}
		transfer = &tr
	}

	addr, err := address.Validate(ta.Address)
	if err != nil {
		return Account{}, errors.Wrap(err, "address")
	}

	return Account{
		Name:      ta.Name,
		Address:   addr,
		Ownership: ownership,
		Transfer:  transfer,
	}, nil
}

type tomlKeypair struct {
	Public  string
	Private string
}

func (tkp tomlKeypair) toKeypair() (Keypair, error) {
	pubBytes, err := base64.StdEncoding.DecodeString(tkp.Public)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unencoding public")
	}
	privBytes, err := base64.StdEncoding.DecodeString(tkp.Private)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unencoding private")
	}
	pub := signature.PublicKey{}
	_, err = pub.UnmarshalMsg(pubBytes)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unmarshaling public")
	}
	priv := signature.PrivateKey{}
	_, err = priv.UnmarshalMsg(privBytes)
	if err != nil {
		return Keypair{}, errors.Wrap(err, "unmarshaling private")
	}
	return Keypair{Public: pub, Private: priv}, nil
}

// A Keypair holds a pair of keys
type Keypair struct {
	Public  signature.PublicKey
	Private signature.PrivateKey
}

func (kp Keypair) toToml() (tomlKeypair, error) {
	pubBytes, err := kp.Public.MarshalMsg(nil)
	if err != nil {
		return tomlKeypair{}, errors.Wrap(err, "marshaling public")
	}
	privBytes, err := kp.Private.MarshalMsg(nil)
	if err != nil {
		return tomlKeypair{}, errors.Wrap(err, "marshaling private")
	}

	return tomlKeypair{
		Public:  base64.StdEncoding.EncodeToString(pubBytes),
		Private: base64.StdEncoding.EncodeToString(privBytes),
	}, nil
}

// String satisfies io.Stringer by writing the public key
func (kp Keypair) String() string {
	return base64.StdEncoding.EncodeToString(kp.Public.Bytes())
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

// DefaultConfig creates a new configuration with the default address
func DefaultConfig() *Config {
	return NewConfig(DefaultAddress)
}

package tool

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/ed25519"
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
	return path.Join(ndauhome, "chaos", "chaostool.toml")
}

// Config represents all data from `chaostool.toml`
type tomlConfig struct {
	Node string

	Identities map[string]tomlIdentity
}

// Identity is a named keypair
type tomlIdentity struct {
	Name       string
	PublicKey  string
	PrivateKey string
}

func (c *tomlConfig) asConfig() (*Config, error) {
	conf := Config{
		Node:       c.Node,
		Identities: make(map[string]Identity, len(c.Identities)),
	}
	for name, tid := range c.Identities {
		public, err := base64.StdEncoding.DecodeString(tid.PublicKey)
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to decode public key for %s: %s",
				name, tid.PublicKey,
			)
		}
		if len(public) != ed25519.PublicKeySize {
			return nil, fmt.Errorf(
				"Wrong public key size for ed25519: have %d, want %d",
				len(public), ed25519.PublicKeySize,
			)
		}
		private, err := base64.StdEncoding.DecodeString(tid.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf(
				"Failed to decode private key for %s: %s",
				name, tid.PrivateKey,
			)
		}
		if len(private) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf(
				"Wrong private key size for ed25519: have %d, want %d",
				len(private), ed25519.PrivateKeySize,
			)
		}

		conf.Identities[name] = Identity{
			Name:       name,
			PublicKey:  public,
			PrivateKey: private,
		}
	}
	return &conf, nil
}

func (c *Config) asTomlConfig() *tomlConfig {
	conf := tomlConfig{
		Node:       c.Node,
		Identities: make(map[string]tomlIdentity, len(c.Identities)),
	}
	for name, id := range c.Identities {
		public := base64.StdEncoding.EncodeToString(id.PublicKey)
		private := base64.StdEncoding.EncodeToString(id.PrivateKey)

		conf.Identities[name] = tomlIdentity{
			Name:       name,
			PublicKey:  public,
			PrivateKey: private,
		}
	}
	return &conf
}

// Config represents all data from `chaostool.toml`
type Config struct {
	Node string

	Identities map[string]Identity
}

// Identity is a named keypair
type Identity struct {
	Name       string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

// Load the current configuration
func Load() (*Config, error) {
	var tconfig tomlConfig
	_, err := toml.DecodeFile(GetConfigPath(), &tconfig)
	if err != nil {
		return nil, err
	}

	config, err := tconfig.asConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Save the current configuration
func (c *Config) Save() error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(c.asTomlConfig()); err != nil {
		return err
	}
	cp := GetConfigPath()
	dir, _ := filepath.Split(cp)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(cp, buf.Bytes(), 0600) // u=rw;go-
}

// NewConfig creates a new configuration with the given address
func NewConfig(node string) *Config {
	return &Config{
		Node:       node,
		Identities: make(map[string]Identity, 0),
	}
}

// DefaultConfig creates a new configuration with the default address
func DefaultConfig() *Config {
	return NewConfig(DefaultAddress)
}

/*
// CreateIdentity with the specified name
func (c *Config) CreateIdentity(name string, out io.Writer) error {
	if _, contained := c.Identities[name]; contained {
		return fmt.Errorf("'%s' already present in Identities map", name)
	}
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	c.Identities[name] = Identity{
		Name:       name,
		PublicKey:  public,
		PrivateKey: private,
	}
	if out != nil {
		EmitIdentityHeader(out)
		EmitIdentity(out, c.Identities[name])
	}
	return nil
}
*/

// ReverseIdentityMap constructs and returns a map from publickey to name
// for configured names.
//
// Note that the map key type is also string. This is a simple cast of
// the public key.
func (c *Config) ReverseIdentityMap() map[string]string {
	out := make(map[string]string)
	for name, id := range c.Identities {
		out[string(id.PublicKey)] = name
	}
	return out
}
